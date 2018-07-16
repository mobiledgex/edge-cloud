package main

import (
	"fmt"
	"log"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"golang.org/x/net/context"
)

func TestLocations(client dme.Match_Engine_ApiClient) {
	var req *dme.Match_Engine_Request
	
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	// Register the client first
	req = new(dme.Match_Engine_Request)
	req.IdType = dme.Match_Engine_Request_IPADDR
	// Should fill out the Id along with carrier and apps details but OK to skip for now
	mstatus, err := client.RegisterClient(ctx, req)
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}
	
	fmt.Println(">>>>>>>Finding Right Locations<<<<<<<<<")
	for _, m := range dmetest.VerifyLocData {
		m.Req.CommCookie = mstatus.CommCookie
		mreply, err := client.VerifyLocation(ctx, &m.Req)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		fmt.Printf("Verify Loc = %f/%f status %d\n",
			m.Req.GpsLocation.Lat, m.Req.GpsLocation.Long,
			mreply.GpsLocationStatus)
	}
}
