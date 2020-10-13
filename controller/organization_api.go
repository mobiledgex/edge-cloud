package main

import (
	"context"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type OrganizationApi struct {
	sync *Sync
}

var organizationApi = OrganizationApi{}

func InitOrganizationApi(sync *Sync) {
	organizationApi.sync = sync
}

func (s *OrganizationApi) OrganizationInUse(ctx context.Context, in *edgeproto.Organization) (*edgeproto.Result, error) {
	usedBy := s.sync.usesOrg(in.Name)
	res := &edgeproto.Result{}
	if len(usedBy) > 0 {
		res.Message = "in use by some " + strings.Join(usedBy, ", ")
		res.Code = 1
	}
	return res, nil
}
