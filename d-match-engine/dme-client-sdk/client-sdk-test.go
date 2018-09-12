package main

import (
	"fmt"
	"log"
	"net"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"golang.org/x/net/context"
)

func FindCloudlets(client dme.Match_Engine_ApiClient) {
	var ipaddr net.IP
	var req *dme.Match_Engine_Request

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	// Register the client first
	req = new(dme.Match_Engine_Request)
	req.IdType = dme.IDTypes_IPADDR
	// Should fill out the Id along with carrier and apps details but OK to skip for now
	mstatus, err := client.RegisterClient(ctx, req)
	if err != nil {
		log.Fatalf("could not register: %v", err)
	}

	fmt.Println(">>>>>>>Finding Right Cloudlets<<<<<<<<<")
	for _, m := range dmetest.FindCloudletData {
		m.Req.SessionCookie = mstatus.SessionCookie
		mreply, err := client.FindCloudlet(ctx, &m.Req)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		ipaddr = mreply.ServiceIp
		fmt.Printf("Found Status %d Loc = %f/%f with Uri/IP %s/%s\n",
			mreply.Status,
			mreply.CloudletLocation.Lat, mreply.CloudletLocation.Long,
			mreply.Uri, ipaddr.String())
	}
}
