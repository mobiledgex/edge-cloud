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

package util

import (
	"fmt"
	"net"
	"os"
)

// Get the external address with port when running under kubernetes.
// This allows horizontally scaled instances to talk to each other within
// a given k8s cluster.
func GetExternalApiAddr(defaultApiAddr string) (string, error) {
	if defaultApiAddr == "" {
		return "", nil
	}
	host, port, err := net.SplitHostPort(defaultApiAddr)
	if err != nil {
		return "", fmt.Errorf("failed to parse api addr %s, %v", defaultApiAddr, err)
	}
	if host == "0.0.0.0" {
		addr, err := ResolveExternalAddr()
		if err == nil {
			defaultApiAddr = addr + ":" + port
		}
	}
	return defaultApiAddr, nil
}

// This is for figuring out the "external" address when
// running under kubernetes, which is really the internal CNI
// address that containers can use to talk to each other.
func ResolveExternalAddr() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}
	return addrs[0], nil
}
