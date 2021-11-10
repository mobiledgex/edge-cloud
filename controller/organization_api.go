package main

import (
	"context"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type OrganizationApi struct {
	all  *AllApis
	sync *Sync
}

func NewOrganizationApi(sync *Sync, all *AllApis) *OrganizationApi {
	organizationApi := OrganizationApi{}
	organizationApi.all = all
	organizationApi.sync = sync
	return &organizationApi
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
