package main

import (
	"fmt"
	"time"
	"log"
	"golang.org/x/net/context"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

func TestLocations(client dme.Match_Engine_ApiClient) {
	var mreq []dme.Match_Engine_Request
		
	mreq = []dme.Match_Engine_Request {
		dme.Match_Engine_Request {
			Carrier: 1,
			//Todo: Check the usage of embedded pointer to Loc in proto files
			GpsLocation: &edgeproto.Loc{Lat: 50.75, Long: 7.9050},
			AppId: 5000,
			DevName: "Untomt",
			AppName: "Untomt",
			AppVers: "1.1",
		},
		dme.Match_Engine_Request {
			Carrier: 1,
			//Todo: Check the usage of embedded pointer to Loc in proto files
			GpsLocation: &edgeproto.Loc{Lat: 52.75, Long: 12.9050},
			AppId: 5005,
			DevName: "Atlantic Labs",
			AppName: "Pillimo-go",
			AppVers: "2.1",
		},
		dme.Match_Engine_Request {
			Carrier: 1,
			//Todo: Check the usage of embedded pointer to Loc in proto files
			GpsLocation: &edgeproto.Loc{Lat: 50.75, Long: 11.9050},
			AppId: 5006,
			DevName: "Atlantic Labs",
			AppName: "HarryPotter-go",
			AppVers: "1.0",
		},
		dme.Match_Engine_Request {
			Carrier: 3,
			//Todo: Check the usage of embedded pointer to Loc in proto files
			GpsLocation: &edgeproto.Loc{Lat: 47.75, Long: 122.9050},
			AppId: 5010,
			DevName: "Ever.AI",
			AppName: "Ever",
			AppVers: "1.7",
		},
	}
	
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	fmt.Println(">>>>>>>Finding Right Cloudlets<<<<<<<<<")
	for _, m := range mreq {
		mreply, err := client.VerifyLocation(ctx, &m)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		fmt.Printf("Verify Loc = %f/%f status %d\n",
			m.GpsLocation.Lat, m.GpsLocation.Long, mreply.GpsLocationStatus);
	}
}
