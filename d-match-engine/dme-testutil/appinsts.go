package dmetest

import (
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type App struct {
	Id           uint64
	Name         string
	Vers         string
	Organization string
}
type Cloudlet struct {
	Id          uint64
	CarrierId   uint64
	CarrierName string
	Name        string
	Uri         string
	Ip          []byte
	Location    dme.Loc
}

var Apps = []App{
	App{
		Id:           5000,
		Name:         "1000realities",
		Vers:         "1.1",
		Organization: "1000realities",
	},
	App{
		Id:           5005,
		Name:         "Pokemon-go",
		Vers:         "2.1",
		Organization: "Niantic Labs",
	},
	App{
		Id:           5006,
		Name:         "HarryPotter-go",
		Vers:         "1.0",
		Organization: "Niantic Labs",
	},
	App{
		Id:           5010,
		Name:         "Ever",
		Vers:         "1.7",
		Organization: "Ever.AI",
	},
	App{
		Id:           5011,
		Name:         "EmptyMatchEngineApp",
		Vers:         "1",
		Organization: "EmptyMatchEngineApp",
	},
	App{
		Id:           5012,
		Name:         cloudcommon.SamsungEnablingLayer,
		Vers:         "1.1",
		Organization: cloudcommon.OrganizationSamsung,
	},
}

var Cloudlets = []Cloudlet{
	Cloudlet{
		Id:          111,
		CarrierId:   1,
		CarrierName: "TDG",
		Name:        "Bonn",
		Uri:         "10.1.10.1",
		Ip:          []byte{10, 1, 10, 1},
		Location:    dme.Loc{Latitude: 50.7374, Longitude: 7.0982},
	},
	Cloudlet{
		Id:          222,
		CarrierId:   1,
		CarrierName: "TDG",
		Name:        "Munich",
		Uri:         "11.1.11.1",
		Ip:          []byte{11, 1, 11, 1},
		Location:    dme.Loc{Latitude: 52.7374, Longitude: 13.4050},
	},
	Cloudlet{
		Id:          333,
		CarrierId:   1,
		CarrierName: "TDG",
		Name:        "Berlin",
		Uri:         "12.1.12.1",
		Ip:          []byte{12, 1, 12, 1},
		Location:    dme.Loc{Latitude: 48.1351, Longitude: 11.5820},
	},
	Cloudlet{
		Id:          444,
		CarrierId:   3,
		CarrierName: "TMUS",
		Name:        "San Francisco",
		Uri:         "13.1.13.1",
		Ip:          []byte{13, 1, 13, 1},
		Location:    dme.Loc{Latitude: 47.6062, Longitude: 122.3321},
	},
}

func MakeAppInst(a *App, c *Cloudlet) *edgeproto.AppInst {
	inst := edgeproto.AppInst{}
	inst.Key.AppKey.Organization = a.Organization
	inst.Key.AppKey.Name = a.Name
	inst.Key.AppKey.Version = a.Vers
	inst.Key.ClusterInstKey.CloudletKey.Organization = c.CarrierName
	inst.Key.ClusterInstKey.CloudletKey.Name = c.Name
	inst.Key.ClusterInstKey.ClusterKey.Name = "testcluster" //TODO - change the testdata to also have clusterInst information
	inst.CloudletLoc = c.Location
	inst.Uri = c.Uri
	inst.State = edgeproto.TrackedState_READY
	inst.HealthCheck = edgeproto.HealthCheck_HEALTH_CHECK_OK // HEALTH_CHECK_OK is not default now
	return &inst
}

func MakeCloudletInfo(c *Cloudlet) *edgeproto.CloudletInfo {
	info := edgeproto.CloudletInfo{}
	info.Key.Organization = c.CarrierName
	info.Key.Name = c.Name
	info.State = edgeproto.CloudletState_CLOUDLET_STATE_READY
	return &info
}

func GenerateApps() []*edgeproto.App {
	apps := make([]*edgeproto.App, 0)
	for _, a := range Apps {
		app := &edgeproto.App{}
		app.Key.Name = a.Name
		app.Key.Organization = a.Organization
		app.Key.Version = a.Vers
		apps = append(apps, app)
	}
	return apps
}

func GenerateAppInsts() []*edgeproto.AppInst {
	insts := make([]*edgeproto.AppInst, 0)
	for _, c := range Cloudlets {
		for _, a := range Apps {
			insts = append(insts, MakeAppInst(&a, &c))
		}
	}
	return insts
}

func GenerateClouldlets() []*edgeproto.CloudletInfo {
	infos := make([]*edgeproto.CloudletInfo, 0)
	for _, c := range Cloudlets {
		infos = append(infos, MakeCloudletInfo(&c))
	}
	return infos

}
