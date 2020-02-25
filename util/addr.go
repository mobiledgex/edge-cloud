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
