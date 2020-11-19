package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CloudletPoolApi struct {
	sync  *Sync
	store edgeproto.CloudletPoolStore
	cache edgeproto.CloudletPoolCache
}

var cloudletPoolApi = CloudletPoolApi{}

func InitCloudletPoolApi(sync *Sync) {
	cloudletPoolApi.sync = sync
	cloudletPoolApi.store = edgeproto.NewCloudletPoolStore(sync.store)
	edgeproto.InitCloudletPoolCache(&cloudletPoolApi.cache)
	sync.RegisterCache(&cloudletPoolApi.cache)
}

func (s *CloudletPoolApi) CreateCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.CloudletPoolAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}

	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		if err := s.checkCloudletsExist(stm, in); err != nil {
			return err
		}
		in.CreatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) DeleteCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if !s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.NotFoundError()
		}
		s.store.STMDel(stm, &in.Key)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) UpdateCloudletPool(ctx context.Context, in *edgeproto.CloudletPool) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := cur.CopyInFields(in)
		if err := cur.Validate(nil); err != nil {
			return err
		}
		if changed == 0 {
			return nil
		}
		if err := s.checkCloudletsExist(stm, &cur); err != nil {
			return err
		}
		cur.UpdatedAt = cloudcommon.TimeToTimestamp(time.Now())
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) checkCloudletsExist(stm concurrency.STM, in *edgeproto.CloudletPool) error {
	notFound := []string{}
	for _, name := range in.Cloudlets {
		key := edgeproto.CloudletKey{
			Name:         name,
			Organization: in.Key.Organization,
		}
		if !cloudletApi.store.STMGet(stm, &key, nil) {
			notFound = append(notFound, name)
		}
	}
	if len(notFound) > 0 {
		return fmt.Errorf("Cloudlets %s not found", strings.Join(notFound, ", "))
	}
	return nil
}

func (s *CloudletPoolApi) ShowCloudletPool(in *edgeproto.CloudletPool, cb edgeproto.CloudletPoolApi_ShowCloudletPoolServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletPool) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletPoolApi) AddCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		for _, name := range cur.Cloudlets {
			if name == in.CloudletName {
				return fmt.Errorf("Cloudlet already part of pool")
			}
		}
		ckey := edgeproto.CloudletKey{
			Name:         in.CloudletName,
			Organization: in.Key.Organization,
		}
		if !cloudletApi.store.STMGet(stm, &ckey, nil) {
			return ckey.NotFoundError()
		}
		cur.Cloudlets = append(cur.Cloudlets, in.CloudletName)
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) RemoveCloudletPoolMember(ctx context.Context, in *edgeproto.CloudletPoolMember) (*edgeproto.Result, error) {
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		cur := edgeproto.CloudletPool{}
		if !s.store.STMGet(stm, &in.Key, &cur) {
			return in.Key.NotFoundError()
		}
		changed := false
		for ii, _ := range cur.Cloudlets {
			if cur.Cloudlets[ii] == in.CloudletName {
				cur.Cloudlets = append(cur.Cloudlets[:ii], cur.Cloudlets[ii+1:]...)
				changed = true
				break
			}
		}
		if !changed {
			return nil
		}
		s.store.STMPut(stm, &cur)
		return nil
	})
	return &edgeproto.Result{}, err
}

func (s *CloudletPoolApi) cloudletDeleted(ctx context.Context, key *edgeproto.CloudletKey) {
	// best effort, no problem if we miss something here.
	toRemove := []edgeproto.CloudletPoolMember{}
	s.cache.Mux.Lock()
	for _, data := range s.cache.Objs {
		if data.Obj.Key.Organization != key.Organization {
			continue
		}
		// we don't check if it's part of the list because
		// the remove func will do that anyway.
		member := edgeproto.CloudletPoolMember{
			Key:          data.Obj.Key,
			CloudletName: key.Name,
		}
		toRemove = append(toRemove, member)
	}
	s.cache.Mux.Unlock()
	for _, member := range toRemove {
		_, err := s.RemoveCloudletPoolMember(ctx, &member)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "failed to remove cloudlet member", "member", member, "err", err)
		}
	}
}
