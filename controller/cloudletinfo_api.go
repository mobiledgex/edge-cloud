package main

import "github.com/mobiledgex/edge-cloud/edgeproto"

type CloudletInfoApi struct {
	sync  *Sync
	store edgeproto.CloudletInfoStore
	cache edgeproto.CloudletInfoCache
}

var cloudletInfoApi = CloudletInfoApi{}

func InitCloudletInfoApi(sync *Sync) {
	cloudletInfoApi.sync = sync
	cloudletInfoApi.store = edgeproto.NewCloudletInfoStore(sync.store)
	edgeproto.InitCloudletInfoCache(&cloudletInfoApi.cache)
	sync.RegisterCache(&cloudletInfoApi.cache)
}

func (s *CloudletInfoApi) ShowCloudletInfo(in *edgeproto.CloudletInfo, cb edgeproto.CloudletInfoApi_ShowCloudletInfoServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.CloudletInfo) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletInfoApi) Update(in *edgeproto.CloudletInfo, notifyId int64) {
	if !cloudletApi.HasKey(in.GetKey()) {
		return
	}
	// for now assume all fields have been specified
	in.Fields = edgeproto.CloudletInfoAllFields
	s.store.Put(in, nil)
}

func (s *CloudletInfoApi) Del(key *edgeproto.CloudletKey, wait func(int64)) {
	in := edgeproto.CloudletInfo{Key: *key}
	s.store.Delete(&in, wait)
}

// Delete from notify just marks the cloudlet offline
func (s *CloudletInfoApi) Delete(in *edgeproto.CloudletInfo, notifyId int64) {
	in.State = edgeproto.CloudletState_Offline
	in.Fields = []string{edgeproto.CloudletInfoFieldState}
	s.store.Put(in, nil)
}

func (s *CloudletInfoApi) Flush(notifyId int64) {
	// mark all cloudlets from the client as offline
	info := edgeproto.CloudletInfo{}
	fields := []string{edgeproto.CloudletInfoFieldState}
	s.cache.Mux.Lock()
	for _, val := range s.cache.Objs {
		if val.NotifyId != notifyId {
			continue
		}
		info.Key = val.Key
		info.State = edgeproto.CloudletState_Offline
		info.Fields = fields
		s.store.Put(&info, nil)
	}
	s.cache.Mux.Unlock()
}
