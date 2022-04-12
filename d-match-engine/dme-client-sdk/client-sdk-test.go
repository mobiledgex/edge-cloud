// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/mobiledgex/edge-cloud/d-match-engine/dme-testutil"
	"golang.org/x/net/context"
)

func FindCloudlets(client dme.MatchEngineApiClient) {
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
		fmt.Printf("Found Status %d Loc = %f/%f with Fqdn %s, ports %v\n",
			mreply.Status,
			mreply.CloudletLocation.Latitude, mreply.CloudletLocation.Longitude,
			mreply.Fqdn, mreply.Ports)
	}
}
