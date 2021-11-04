package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type DeviceApi struct {
	all   *AllApis
	sync  *Sync
	store edgeproto.DeviceStore
	cache edgeproto.DeviceCache
}

func NewDeviceApi(sync *Sync, all *AllApis) *DeviceApi {
	deviceApi := DeviceApi{}
	deviceApi.all = all
	deviceApi.sync = sync
	deviceApi.store = edgeproto.NewDeviceStore(sync.store)
	edgeproto.InitDeviceCache(&deviceApi.cache)
	sync.RegisterCache(&deviceApi.cache)
	return &deviceApi
}

func (s *DeviceApi) ShowDevice(in *edgeproto.Device, cb edgeproto.DeviceApi_ShowDeviceServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Device) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *DeviceApi) CreateDevice(ctx context.Context, in *edgeproto.Device) (*edgeproto.Result, error) {
	// Unsupported - use InjectDevice instead
	return &edgeproto.Result{}, fmt.Errorf("Use InjectDevice instead")
}

func (s *DeviceApi) InjectDevice(ctx context.Context, in *edgeproto.Device) (*edgeproto.Result, error) {
	if err := in.Validate(edgeproto.DeviceAllFieldsMap); err != nil {
		return &edgeproto.Result{}, err
	}
	err := s.sync.ApplySTMWait(ctx, func(stm concurrency.STM) error {
		if s.store.STMGet(stm, &in.Key, nil) {
			return in.Key.ExistsError()
		}
		s.store.STMPut(stm, in)
		return nil
	})
	return &edgeproto.Result{}, err
}

// This api deletes the device from the controller cache
func (s *DeviceApi) EvictDevice(ctx context.Context, in *edgeproto.Device) (*edgeproto.Result, error) {
	return s.store.Delete(ctx, in, s.sync.syncWait)
}

// Show devices that showed up in this timestamp
func (s *DeviceApi) ShowDeviceReport(in *edgeproto.DeviceReport, cb edgeproto.DeviceApi_ShowDeviceReportServer) error {
	filter := edgeproto.Device{
		Key: in.Key,
	}
	err := s.cache.Show(&filter, func(obj *edgeproto.Device) error {
		if (in.Begin == nil || obj.FirstSeen.Compare(in.Begin) >= 0) &&
			(in.End == nil || obj.FirstSeen.Compare(in.End) <= 0) {
			err := cb.Send(obj)
			return err
		}
		return nil
	})
	return err
}

// Does the same as create - once the device is stored it's there forever
func (s *DeviceApi) Update(ctx context.Context, in *edgeproto.Device, rev int64) {
	if !s.HasDevice(&in.Key) {
		if _, err := s.InjectDevice(ctx, in); err != nil {
			log.SpanLog(ctx, log.DebugLevelApi, "Failed to add platform client",
				"client", in)
		}
	}
}

func (s *DeviceApi) Delete(ctx context.Context, in *edgeproto.Device, rev int64) {
	// Eviction is the only way to get rid of the device
}

func (s *DeviceApi) Flush(ctx context.Context, notifyId int64) {
	// Flush does not delete anything in the controller cache
}

func (s *DeviceApi) Prune(ctx context.Context, keys map[edgeproto.DeviceKey]struct{}) {}

func (s *DeviceApi) HasDevice(key *edgeproto.DeviceKey) bool {
	return s.cache.HasKey(key)
}
