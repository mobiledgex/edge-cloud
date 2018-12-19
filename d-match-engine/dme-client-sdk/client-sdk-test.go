package main

import (
	"fmt"
	"log"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"golang.org/x/net/context"
)

func FindCloudlets(client dme.Match_Engine_ApiClient) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	fmt.Println(">>>>>>>Finding Right Cloudlets<<<<<<<<<")
	for _, m := range dmetest.FindCloudletData {
		mstatus, err := client.RegisterClient(ctx, &m.Reg)
		if err != nil {
			log.Fatalf("could not register: %v", err)
		}

		m.Req.SessionCookie = mstatus.SessionCookie
		mreply, err := client.FindCloudlet(ctx, &m.Req)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		fmt.Printf("Found Status %d Loc = %f/%f with FQDN %s, ports %v\n",
			mreply.Status,
			mreply.CloudletLocation.Latitude, mreply.CloudletLocation.Longitude,
			mreply.FQDN, mreply.Ports)
	}
}
