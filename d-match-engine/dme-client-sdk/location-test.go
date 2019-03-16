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
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	fmt.Println(">>>>>>>Finding Right Locations<<<<<<<<<")
	for _, m := range dmetest.VerifyLocData {
		// Register the client first
		mstatus, err := client.RegisterClient(ctx, &m.Reg)
		if err != nil {
			log.Fatalf("could not register: %v", err)
		}
		m.Req.SessionCookie = mstatus.SessionCookie
		mreply, err := client.VerifyLocation(ctx, &m.Req)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		fmt.Printf("Verify Loc = %f/%f status %d\n",
			m.Req.GPSLocation.Latitude, m.Req.GPSLocation.Longitude,
			mreply.GPS_Location_Status)
	}
}
