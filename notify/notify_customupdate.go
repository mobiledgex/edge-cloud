package notify

import (
	"context"

	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

// Customize functions are used to filter sending of data
// to the CRM by sending only objects related to the CloudletKey.
// The remote initially tells us it wants cloudletKey filtering.
// If so, none of the below objects are sent until we receive a
// cloudletkey via CloudletInfo. Then further updates (sends) are
// filtered by cloudletkey(s).

func (s *AppSend) UpdateOk(ctx context.Context, key *edgeproto.AppKey) bool {
	if s.sendrecv.filterCloudletKeys {
		// Apps gets sent with AppInst
		return false
	}
	return true
}

func (s *AppInstSend) UpdateOk(ctx context.Context, key *edgeproto.AppInstKey) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.cloudletReady {
			return false
		}
		if !s.sendrecv.hasCloudletKey(&key.ClusterInstKey.CloudletKey) {
			return false
		}
		// also trigger sending app
		if s.sendrecv.appSend != nil {
			s.sendrecv.appSend.updateInternal(ctx, &key.AppKey, 0)
		}
	}
	return true
}

func (s *CloudletSend) UpdateOk(ctx context.Context, key *edgeproto.CloudletKey) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.hasCloudletKey(key) {
			return false
		}
		cloudlet := edgeproto.Cloudlet{}
		var modRev int64
		// also trigger send of vmpool object
		if s.handler.GetWithRev(key, &cloudlet, &modRev) && cloudlet.VmPool != "" {
			s.sendrecv.vmPoolSend.updateInternal(ctx, &edgeproto.VMPoolKey{
				Name:         cloudlet.VmPool,
				Organization: key.Organization,
			}, 0)
		}
	}
	return true
}

func (s *ClusterInstSend) UpdateOk(ctx context.Context, key *edgeproto.ClusterInstKey) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.cloudletReady {
			return false
		}
		if !s.sendrecv.hasCloudletKey(&key.CloudletKey) {
			return false
		}
	}
	return true
}

func (s *ExecRequestSend) UpdateOk(ctx context.Context, msg *edgeproto.ExecRequest) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.cloudletReady {
			return false
		}
		if !s.sendrecv.hasCloudletKey(&msg.AppInstKey.ClusterInstKey.CloudletKey) {
			return false
		}
	}
	return true
}

func (s *VMPoolSend) UpdateOk(ctx context.Context, key *edgeproto.VMPoolKey) bool {
	if s.sendrecv.filterCloudletKeys {
		for cKey, _ := range s.sendrecv.cloudletKeys {
			cloudlet := edgeproto.Cloudlet{}
			var modRev int64
			if cKey.Organization != key.Organization {
				continue
			}
			if s.sendrecv.cloudletSend.handler.GetWithRev(&cKey, &cloudlet, &modRev) {
				if cloudlet.VmPool != key.Name {
					continue
				}
				return true
			}
		}
		return false
	}
	return true
}

func (s *AppSend) UpdateAllOk() bool {
	return !s.sendrecv.filterCloudletKeys
}

func (s *AppInstSend) UpdateAllOk() bool {
	return !s.sendrecv.filterCloudletKeys
}

func (s *CloudletSend) UpdateAllOk() bool {
	return !s.sendrecv.filterCloudletKeys
}

func (s *ClusterInstSend) UpdateAllOk() bool {
	return !s.sendrecv.filterCloudletKeys
}

func (s *VMPoolSend) UpdateAllOk() bool {
	return !s.sendrecv.filterCloudletKeys
}

func (s *CloudletInfoRecv) RecvHook(ctx context.Context, notice *edgeproto.Notice, buf *edgeproto.CloudletInfo, peerAddr string) {
	log.SpanLog(ctx, log.DebugLevelNotify, "CloudletInfo RecvHook", "key", buf.Key, "state", buf.State)

	if !s.sendrecv.filterCloudletKeys {
		return
	}

	// set filter to allow sending of cloudlet data
	s.sendrecv.updateCloudletKey(notice.Action, &buf.Key)

	if notice.Action == edgeproto.NoticeAction_UPDATE {
		if buf.State == dmeproto.CloudletState_CLOUDLET_STATE_READY ||
			buf.State == dmeproto.CloudletState_CLOUDLET_STATE_UPGRADE ||
			buf.State == dmeproto.CloudletState_CLOUDLET_STATE_NEED_SYNC ||
			buf.State == dmeproto.CloudletState_CLOUDLET_STATE_INIT {
			// trigger send of cloudlet details to cloudlet
			if s.sendrecv.cloudletSend != nil {
				log.SpanLog(ctx, log.DebugLevelNotify, "CloudletInfo recv hook, send Cloudlet", "key", buf.Key, "state", buf.State)
				s.sendrecv.cloudletSend.Update(ctx, &buf.Key, nil, 0)
			}
		}
		if buf.State == dmeproto.CloudletState_CLOUDLET_STATE_READY || buf.State == dmeproto.CloudletState_CLOUDLET_STATE_NEED_SYNC && !buf.ControllerCacheReceived {
			log.SpanLog(ctx, log.DebugLevelNotify, "CloudletInfo recv hook read, send all filtered data", "key", buf.Key)
			// allow all filtered objects to be sent
			s.sendrecv.cloudletReady = true

			// trigger send of all objects related to cloudlet
			// In case of cloudlet upgrade, Check if READY is
			// received from the appropriate cloudlet
			cloudlet := edgeproto.Cloudlet{}
			var modRev int64
			if buf.ContainerVersion != "" && s.sendrecv.cloudletSend != nil {
				if s.sendrecv.cloudletSend.handler.GetWithRev(&buf.Key, &cloudlet, &modRev) &&
					(cloudlet.State == edgeproto.TrackedState_UPDATE_REQUESTED ||
						cloudlet.State == edgeproto.TrackedState_UPDATING) &&
					cloudlet.ContainerVersion != buf.ContainerVersion {
					return
				}
			}

			// Post cloudlet upgrade, when CLOUDLET_STATE_READY state
			// is seen from upgraded CRM, then following will trigger
			// send of all objects (which includes objects missed
			// during upgrade)
			if s.sendrecv.clusterInstSend != nil {
				clusterInsts := make(map[edgeproto.ClusterInstKey]int64)
				s.sendrecv.clusterInstSend.handler.GetForCloudlet(&buf.Key, func(key *edgeproto.ClusterInstKey, modRev int64) {
					clusterInsts[*key] = modRev
				})
				for k, modRev := range clusterInsts {
					s.sendrecv.clusterInstSend.Update(ctx, &k, nil, modRev)
				}
			}
			if s.sendrecv.appInstSend != nil {
				appInsts := make(map[edgeproto.AppInstKey]int64)
				s.sendrecv.appInstSend.handler.GetForCloudlet(&buf.Key, func(key *edgeproto.AppInstKey, modRev int64) {
					appInsts[*key] = modRev
				})
				for k, modRev := range appInsts {
					s.sendrecv.appInstSend.Update(ctx, &k, nil, modRev)
				}
			}
			s.sendrecv.triggerSendAllEnd()

		}
	}
}

func (s *CloudletRecv) RecvHook(ctx context.Context, notice *edgeproto.Notice, buf *edgeproto.Cloudlet, perrAddr string) {
	// register cloudlet key on sendrecv for CRM, otherwise the
	// ExecRequest messages it tries to send back to the controller
	// will get filtered by UpdateOk above.
	s.sendrecv.updateCloudletKey(notice.Action, &buf.Key)
	s.sendrecv.cloudletReady = true
}
