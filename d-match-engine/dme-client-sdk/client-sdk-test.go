package main

import (
	"fmt"
	"time"
	"log"
	"net"
	"golang.org/x/net/context"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

func find_cloudlets(client dme.Match_Engine_ApiClient) {
	var mreq []dme.Match_Engine_Request
	var ipaddr net.IP
	//var mreply dme.Match_Engine_Reply
	
	mreq = []dme.Match_Engine_Request {
		dme.Match_Engine_Request {
			Carrier: 1,
			GpsLocation: &edgeproto.Loc{Lat: 50.75, Long: 7.9050},
			AppId: 5000,
		},
		dme.Match_Engine_Request {
			Carrier: 1,
			GpsLocation: &edgeproto.Loc{Lat: 52.75, Long: 12.9050},
			AppId: 5005,
		},
		dme.Match_Engine_Request {
			Carrier: 1,
			GpsLocation: &edgeproto.Loc{Lat: 50.75, Long: 11.9050},
			AppId: 5006,
		},
		dme.Match_Engine_Request {
			Carrier: 3,
			GpsLocation: &edgeproto.Loc{Lat: 47.75, Long: 122.9050},
			AppId: 5010,
		},
	}
	
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	fmt.Println(">>>>>>>Finding Right Cloudlets<<<<<<<<<")
	for _, m := range mreq {
		mreply, err := client.FindCloudlet(ctx, &m)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		ipaddr = mreply.ServiceIp
		fmt.Printf("Found Loc = %f/%f with IP %s\n",
			mreply.CloudletLocation.Lat, mreply.CloudletLocation.Long, ipaddr.String());
	}
}
