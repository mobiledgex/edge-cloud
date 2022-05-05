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

	dme "github.com/edgexr/edge-cloud/d-match-engine/dme-proto"
	dmetest "github.com/edgexr/edge-cloud/d-match-engine/dme-testutil"
	"golang.org/x/net/context"
)

func TestLocations(client dme.MatchEngineApiClient) {
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
			m.Req.GpsLocation.Latitude, m.Req.GpsLocation.Longitude,
			mreply.GpsLocationStatus)
	}
}
