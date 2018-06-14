package main

import (
	"fmt"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

func populate_tbl() {
	var apps []app
	var cloudlets []cloudlet

	apps = []app{
		app{
			id:        5000,
			name:      "1000realities",
			vers:      "1.1",
			developer: "1000realities",
		},
		app{
			id:        5005,
			name:      "Pokemon-go",
			vers:      "2.1",
			developer: "Niantic Labs",
		},
		app{
			id:        5006,
			name:      "HarryPotter-go",
			vers:      "1.0",
			developer: "Niantic Labs",
		},
		app{
			id:        5010,
			name:      "Ever",
			vers:      "1.7",
			developer: "Ever.AI",
		},
		app{
			id:        5011,
			name:      "EmptyMatchEngineApp",
			vers:      "1",
			developer: "EmptyMatchEngineApp",
		},
	}
	cloudlets = []cloudlet{
		cloudlet{
			id:          111,
			carrierId:   1,
			carrierName: "TDG",
			accessIp:    []byte{10, 1, 10, 1},
			location:    dme.Loc{Lat: 50.7374, Long: 7.0982},
		},
		cloudlet{
			id:          222,
			carrierId:   1,
			carrierName: "TDG",
			accessIp:    []byte{11, 1, 11, 1},
			location:    dme.Loc{Lat: 52.7374, Long: 13.4050},
		},
		cloudlet{
			id:          333,
			carrierId:   1,
			carrierName: "TDG",
			accessIp:    []byte{12, 1, 12, 1},
			location:    dme.Loc{Lat: 48.1351, Long: 11.5820},
		},
		cloudlet{
			id:          444,
			carrierId:   3,
			carrierName: "TMUS",
			accessIp:    []byte{13, 1, 13, 1},
			location:    dme.Loc{Lat: 47.6062, Long: 122.3321},
		},
	}

	for _, c := range cloudlets {
		fmt.Printf("Key = %d, name = %s\n", c.id, c.carrierName)
		for _, a := range apps {
			add_app(&a, &c)
		}
	}
}
