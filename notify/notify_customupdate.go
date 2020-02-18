package notify

import (
	"context"

	"github.com/mobiledgex/edge-cloud/edgeproto"
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
		if !s.sendrecv.hasCloudletKey(&key.ClusterInstKey.CloudletKey) {
			return false
		}
		// also trigger sending app
		if s.sendrecv.appSend != nil {
			s.sendrecv.appSend.updateInternal(ctx, &key.AppKey)
		}
	}
	return true
}

func (s *CloudletSend) UpdateOk(ctx context.Context, key *edgeproto.CloudletKey) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.hasCloudletKey(key) {
			return false
		}
	}
	return true
}

func (s *ClusterInstSend) UpdateOk(ctx context.Context, key *edgeproto.ClusterInstKey) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.hasCloudletKey(&key.CloudletKey) {
			return false
		}
	}
	return true
}

func (s *ExecRequestSend) UpdateOk(ctx context.Context, msg *edgeproto.ExecRequest) bool {
	if s.sendrecv.filterCloudletKeys {
		if !s.sendrecv.hasCloudletKey(&msg.AppInstKey.ClusterInstKey.CloudletKey) {
			return false
		}
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

func (s *CloudletInfoRecv) RecvHook(ctx context.Context, notice *edgeproto.Notice, buf *edgeproto.CloudletInfo, peerAddr string) {
	if !s.sendrecv.filterCloudletKeys {
		return
	}

	s.sendrecv.updateCloudletKey(notice.Action, &buf.Key)

	if notice.Action == edgeproto.NoticeAction_UPDATE {
		if buf.State == edgeproto.CloudletState_CLOUDLET_STATE_READY ||
			buf.State == edgeproto.CloudletState_CLOUDLET_STATE_UPGRADE ||
			buf.State == edgeproto.CloudletState_CLOUDLET_STATE_INIT {
			// trigger send of cloudlet details to cloudlet
			if s.sendrecv.cloudletSend != nil {
				s.sendrecv.cloudletSend.Update(ctx, &buf.Key, nil)
			}
		}
		if buf.State == edgeproto.CloudletState_CLOUDLET_STATE_READY {
			// trigger send of all objects related to cloudlet

			// In case of cloudlet upgrade, Check if READY is recieved from
			// appropriate cloudlet
			cloudlet := edgeproto.Cloudlet{}
			if buf.ContainerVersion != "" && s.sendrecv.cloudletSend != nil {
				if s.sendrecv.cloudletSend.handler.Get(&buf.Key, &cloudlet) &&
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
				clusterInsts := make(map[edgeproto.ClusterInstKey]struct{})
				s.sendrecv.clusterInstSend.handler.GetForCloudlet(&buf.Key, clusterInsts)
				for k, _ := range clusterInsts {
					s.sendrecv.clusterInstSend.Update(ctx, &k, nil)
				}
			}
			if s.sendrecv.appInstSend != nil {
				appInsts := make(map[edgeproto.AppInstKey]struct{})
				s.sendrecv.appInstSend.handler.GetForCloudlet(&buf.Key, appInsts)
				for k, _ := range appInsts {
					s.sendrecv.appInstSend.Update(ctx, &k, nil)
				}
			}
		}
	}
}

func (s *CloudletRecv) RecvHook(ctx context.Context, notice *edgeproto.Notice, buf *edgeproto.Cloudlet, perrAddr string) {
	// register cloudlet key on sendrecv for CRM, otherwise the
	// ExecRequest messages it tries to send back to the controller
	// will get filtered by UpdateOk above.
	s.sendrecv.updateCloudletKey(notice.Action, &buf.Key)
}
