package dmetest

import "github.com/mobiledgex/edge-cloud/edgeproto"
import dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"

type App struct {
	Id        uint64
	Name      string
	Vers      string
	Developer string
}
type Cloudlet struct {
	Id          uint64
	CarrierId   uint64
	CarrierName string
	Name        string
	Uri         string
	Location    dme.Loc
}

var Apps = []App{
	App{
		Id:        5000,
		Name:      "1000realities",
		Vers:      "1.1",
		Developer: "1000realities",
	},
	App{
		Id:        5005,
		Name:      "Pokemon-go",
		Vers:      "2.1",
		Developer: "Niantic Labs",
	},
	App{
		Id:        5006,
		Name:      "HarryPotter-go",
		Vers:      "1.0",
		Developer: "Niantic Labs",
	},
	App{
		Id:        5010,
		Name:      "Ever",
		Vers:      "1.7",
		Developer: "Ever.AI",
	},
	App{
		Id:        5011,
		Name:      "EmptyMatchEngineApp",
		Vers:      "1",
		Developer: "EmptyMatchEngineApp",
	},
}

var Cloudlets = []Cloudlet{
	Cloudlet{
		Id:          111,
		CarrierId:   1,
		CarrierName: "TDG",
		Name:        "Bonn",
		Uri:         "10.1.10.1",
		Location:    dme.Loc{Lat: 50.7374, Long: 7.0982},
	},
	Cloudlet{
		Id:          222,
		CarrierId:   1,
		CarrierName: "TDG",
		Name:        "Munich",
		Uri:         "11.1.11.1",
		Location:    dme.Loc{Lat: 52.7374, Long: 13.4050},
	},
	Cloudlet{
		Id:          333,
		CarrierId:   1,
		CarrierName: "TDG",
		Name:        "Berlin",
		Uri:         "12.1.12.1",
		Location:    dme.Loc{Lat: 48.1351, Long: 11.5820},
	},
	Cloudlet{
		Id:          444,
		CarrierId:   3,
		CarrierName: "TMUS",
		Name:        "San Francisco",
		Uri:         "13.1.13.1",
		Location:    dme.Loc{Lat: 47.6062, Long: 122.3321},
	},
}

func MakeAppInst(a *App, c *Cloudlet) *edgeproto.AppInst {
	inst := edgeproto.AppInst{}
	inst.Key.AppKey.DeveloperKey.Name = a.Developer
	inst.Key.AppKey.Name = a.Name
	inst.Key.AppKey.Version = a.Vers
	inst.Key.CloudletKey.OperatorKey.Name = c.CarrierName
	inst.Key.CloudletKey.Name = c.Name
	inst.CloudletLoc = c.Location
	inst.Uri = c.Uri
	return &inst
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
