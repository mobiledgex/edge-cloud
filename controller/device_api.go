package main

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DeviceApi struct {
	sync  *Sync
	store edgeproto.DeviceStore
	cache edgeproto.DeviceCache
}

var deviceApi = DeviceApi{}

func (s *DeviceApi) ShowDevice(in *edgeproto.Device, cb edgeproto.DeviceApi_ShowDeviceServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Device) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *DeviceApi) Update(ctx context.Context, in *edgeproto.Device, rev int64) {
	if !s.HasDevice(&in.Key) {
		s.cache.Update(ctx, in, rev)
	}
}

func (s *DeviceApi) Delete(ctx context.Context, in *edgeproto.Device, rev int64) {
	s.cache.Delete(ctx, in, rev)
}

func (s *DeviceApi) Flush(ctx context.Context, notifyId int64) {
	s.cache.Flush(ctx, notifyId)
}

func (s *DeviceApi) Prune(ctx context.Context, keys map[edgeproto.DeviceKey]struct{}) {}

func (s *DeviceApi) HasDevice(key *edgeproto.DeviceKey) bool {
	return s.cache.HasKey(key)
}
