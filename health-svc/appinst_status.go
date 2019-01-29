package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func AppInstGetStatus(inst *edgeproto.AppInst, timeout time.Duration) string {
	statuses := []string{}
	for _, port := range inst.MappedPorts {
		uri := fmt.Sprintf("%s%s:%d", port.FQDNPrefix,
			inst.Uri, port.PublicPort)
		portStatus := ""
		switch port.Proto {
		case dme.LProto_LProtoHTTP:
			// Here's an example of a GET request check.
			/*
				uri := "http://" + uri + port.PublicPath
				client := http.Client{
					Timeout: timeout,
				}
				resp, err := client.Get(uri)
				if err != nil {
					portStatus = err.Error()
				} else if resp.StatusCode != http.StatusOK {
					portStatus = fmt.Sprintf("GET %s: %s", uri,
						resp.Status)
				}
			*/
			// However, we don't really know if the app
			// is expecting a GET or POST, or what to expect
			// in response. We'd need specific info about the
			// the app to be able to validate an http request.
		case dme.LProto_LProtoTCP:
			conn, err := net.DialTimeout("tcp", uri, timeout)
			if err != nil {
				portStatus = err.Error()
			} else {
				conn.Close()
			}
		case dme.LProto_LProtoUDP:
			// No real way to test udp. A valid udp listener
			// that doesn't respond looks the same as a
			// missing listener or black hole, due to no
			// handshakes for udp.
			// So testing udp would be application-specific.
		}
		if portStatus != "" {
			statuses = append(statuses, portStatus)
		}
	}
	status := "ok"
	if len(statuses) > 0 {
		status = strings.Join(statuses, ", ")
	}
	return status
}
