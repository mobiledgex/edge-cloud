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

package cloudflaremgmt

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/edgexr/edge-cloud/log"
)

var mexTestInfra = os.Getenv("MEX_TEST_INFRA")

var user, key, domain string
var testapi *cloudflare.API

func TestInit(t *testing.T) {
	if mexTestInfra == "" {
		return
	}
	log.SetDebugLevel(log.DebugLevelInfra)

	user = os.Getenv("CF_USER")
	key = os.Getenv("CF_KEY")
	domain = os.Getenv("CF_TEST_DOMAIN")

	if user == "" {
		t.Errorf("missing CF_USER environment variable")
	}
	if key == "" {
		t.Errorf("missing CF_KEY environment variable")
	}
	if domain == "" {
		t.Errorf("missing CF_TEST_DOMAIN environment variable")
	}

	api, err := cloudflare.New(key, user)
	if err != nil {
		t.Errorf("cannot initialize API, %v", err)
	}
	testapi = api
}

func TestGetDNSRecords(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelInfra)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	if mexTestInfra == "" {
		return
	}
	// XXX not sure if domain is zone or name, but guessing zone
	recs, err := GetDNSRecords(ctx, testapi, domain, "")
	if err != nil {
		t.Errorf("can not get dns records for %s, %v", domain, err)
	}

	for _, rec := range recs {
		fmt.Println(rec)
	}
}

func TestCreateDNSRecord(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelInfra)
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	if mexTestInfra == "" {
		return
	}
	cname := "name-test-1"
	err := CreateDNSRecord(ctx, testapi, domain, cname, "cname", domain, 1, false)
	if err != nil {
		t.Errorf("cannot create DNS record, %v", err)
	}

	//try to recreate -- should produce error
	err = CreateDNSRecord(ctx, testapi, domain, cname, "cname", domain, 1, false)
	if err == nil {
		t.Errorf("should have failed")
	}

	recs, err := GetDNSRecords(ctx, testapi, domain, "")
	if err != nil {
		t.Errorf("can not get dns records for %s, %v", domain, err)
	}

	for _, rec := range recs {
		if strings.HasPrefix(rec.Name, cname) {
			err = DeleteDNSRecord(ctx, testapi, domain, rec.ID)
			if err != nil {
				t.Errorf("cannot delete dns record id %s zone %s, %v", rec.ID, domain, err)
			}
		}
	}
}
