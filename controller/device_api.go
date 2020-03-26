package main

import (
	"context"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type DeviceApi struct {
	sync  *Sync
	store edgeproto.DeviceStore
	cache edgeproto.DeviceCache
}

var deviceApi = DeviceApi{}

func InitDeviceApi(sync *Sync) {
	deviceApi.sync = sync
	deviceApi.store = edgeproto.NewDeviceStore(sync.store)
	edgeproto.InitDeviceCache(&deviceApi.cache)
	sync.RegisterCache(&deviceApi.cache)
}

func (s *DeviceApi) ShowDevice(in *edgeproto.Device, cb edgeproto.DeviceApi_ShowDeviceServer) error {
	filter := edgeproto.Device{}
	err := s.cache.Show(&filter, func(obj *edgeproto.Device) error {
		log.SpanLog(cb.Context(), log.DebugLevelApi, "Showing client", "client", obj)
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *DeviceApi) CreateDevice(ctx context.Context, in *edgeproto.Device) (*edgeproto.Result, error) {
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

// Does the same as create - once the device is stored it's there forever
func (s *DeviceApi) Update(ctx context.Context, in *edgeproto.Device, rev int64) {
	if err := in.Validate(edgeproto.DeviceAllFieldsMap); err != nil {
		return
	}

	if !s.HasDevice(&in.Key) {
		if _, err := s.CreateDevice(ctx, in); err != nil {
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
