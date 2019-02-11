package main

import (
	"errors"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

type CloudletApi struct {
	sync  *Sync
	store edgeproto.CloudletStore
	cache edgeproto.CloudletCache
}

var cloudletApi = CloudletApi{}

func InitCloudletApi(sync *Sync) {
	cloudletApi.sync = sync
	cloudletApi.store = edgeproto.NewCloudletStore(sync.store)
	edgeproto.InitCloudletCache(&cloudletApi.cache)
	sync.RegisterCache(&cloudletApi.cache)
}

func (s *CloudletApi) GetAllKeys(keys map[edgeproto.CloudletKey]struct{}) {
	s.cache.GetAllKeys(keys)
}

func (s *CloudletApi) Get(key *edgeproto.CloudletKey, buf *edgeproto.Cloudlet) bool {
	return s.cache.Get(key, buf)
}

func (s *CloudletApi) HasKey(key *edgeproto.CloudletKey) bool {
	return s.cache.HasKey(key)
}

func (s *CloudletApi) UsesOperator(in *edgeproto.OperatorKey) bool {
	s.cache.Mux.Lock()
	defer s.cache.Mux.Unlock()
	for key, _ := range s.cache.Objs {
		if key.OperatorKey.Matches(in) {
			return true
		}
	}
	return false
}

func (s *CloudletApi) CreateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_CreateCloudletServer) error {
	if in.IpSupport == edgeproto.IpSupport_IpSupportUnknown {
		in.IpSupport = edgeproto.IpSupport_IpSupportDynamic
	}
	// TODO: support static IP assignment.
	if in.IpSupport != edgeproto.IpSupport_IpSupportDynamic {
		return errors.New("Only dynamic IPs are supported currently")
	}
	if in.IpSupport == edgeproto.IpSupport_IpSupportStatic {
		// TODO: Validate static ips
	} else {
		// dynamic
		if in.NumDynamicIps < 1 {
			return errors.New("Must specify at least one dynamic public IP available")
		}
	}
	if in.Location.Latitude == 0 && in.Location.Longitude == 0 {
		// user forgot to specify location
		return errors.New("location is missing; 0,0 is not a valid location")
	}
	_, err := s.store.Create(in, s.sync.syncWait)
	return err
}

func (s *CloudletApi) UpdateCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_UpdateCloudletServer) error {
	fmap := edgeproto.MakeFieldMap(in.Fields)
	if _, found := fmap[edgeproto.CloudletFieldNumDynamicIps]; found {
		staticSet := false
		if _, staticFound := fmap[edgeproto.CloudletFieldIpSupport]; staticFound {
			if in.IpSupport == edgeproto.IpSupport_IpSupportStatic {
				staticSet = true
			}
		}
		if in.NumDynamicIps < 1 && !staticSet {
			return errors.New("Cannot specify less than one dynamic IP unless Ip Support Static is specified")
		}
	}

	_, err := s.store.Update(in, s.sync.syncWait)

	// after the cloudlet change is committed, if the location changed,
	// update app insts as well.
	s.UpdateAppInstLocations(in)
	return err
}

func (s *CloudletApi) DeleteCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_DeleteCloudletServer) error {
	dynInsts := make(map[edgeproto.AppInstKey]struct{})
	if appInstApi.UsesCloudlet(&in.Key, dynInsts) {
		// disallow delete if static instances are present
		return errors.New("Cloudlet in use by static Application Instance")
	}
	clDynInsts := make(map[edgeproto.ClusterInstKey]struct{})
	if clusterInstApi.UsesCloudlet(&in.Key, clDynInsts) {
		return errors.New("Cloudlet in use by static Cluster Instance")
	}
	_, err := s.store.Delete(in, s.sync.syncWait)
	// also delete associated info
	// Note: don't delete cloudletinfo, that will get deleted once CRM
	// disconnects. Otherwise if admin deletes/recreates Cloudlet with
	// CRM connected the whole time, we will end up without cloudletInfo.
	cloudletRefsApi.Delete(&in.Key, s.sync.syncWait)
	// also delete dynamic instances
	if len(dynInsts) > 0 {
		// delete dynamic instances
		for key, _ := range dynInsts {
			appInst := edgeproto.AppInst{Key: key}
			derr := appInstApi.deleteAppInstInternal(DefCallContext(), &appInst, cb)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic app inst",
					"key", key, "err", derr)
			}
		}
	}
	if len(clDynInsts) > 0 {
		for key, _ := range clDynInsts {
			clInst := edgeproto.ClusterInst{Key: key}
			derr := clusterInstApi.deleteClusterInstInternal(DefCallContext(), &clInst, cb)
			if derr != nil {
				log.DebugLog(log.DebugLevelApi,
					"Failed to delete dynamic cluster inst",
					"key", key, "err", derr)
			}
		}
	}
	return err
}

func (s *CloudletApi) ShowCloudlet(in *edgeproto.Cloudlet, cb edgeproto.CloudletApi_ShowCloudletServer) error {
	err := s.cache.Show(in, func(obj *edgeproto.Cloudlet) error {
		err := cb.Send(obj)
		return err
	})
	return err
}

func (s *CloudletApi) UpdateAppInstLocations(in *edgeproto.Cloudlet) {
	fmap := edgeproto.MakeFieldMap(in.Fields)
	if _, found := fmap[edgeproto.CloudletFieldLocation]; !found {
		// no location fields updated
		return
	}

	// find all appinsts associated with the cloudlet
	keys := make([]edgeproto.AppInstKey, 0)
	appInstApi.cache.Mux.Lock()
	for _, inst := range appInstApi.cache.Objs {
		if inst.Key.CloudletKey.Matches(&in.Key) {
			keys = append(keys, inst.Key)
		}
	}
	appInstApi.cache.Mux.Unlock()

	first := true
	inst := edgeproto.AppInst{}
	for ii, _ := range keys {
		inst.Key = keys[ii]
		if first {
			inst.Fields = make([]string, 0)
		}
		if _, found := fmap[edgeproto.CloudletFieldLocationLatitude]; found {
			inst.CloudletLoc.Latitude = in.Location.Latitude
			if first {
				inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLatitude)
			}
		}
		if _, found := fmap[edgeproto.CloudletFieldLocationLongitude]; found {
			inst.CloudletLoc.Longitude = in.Location.Longitude
			if first {
				inst.Fields = append(inst.Fields, edgeproto.AppInstFieldCloudletLocLongitude)
			}
		}
		if first {
			if len(inst.Fields) == 0 {
				break
			}
			first = false
		}

		err := appInstApi.updateAppInstInternal(DefCallContext(), &inst, nil)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "Update AppInst Location",
				"inst", inst, "err", err)
		}
	}
}
