package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/require"
)

func TestGetStatus(t *testing.T) {
	timeout := 20 * time.Millisecond

	// http test listener
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.Nil(t, err)
	require.Equal(t, "127.0.0.1", u.Hostname())

	// tcp test listener
	ltcp, err := net.Listen("tcp", "127.0.0.1:44444")
	require.Nil(t, err)
	defer ltcp.Close()

	// udp test listener (not really needed since we can't test)
	ludp, err := net.ListenPacket("udp", "127.0.0.1:44444")
	require.Nil(t, err)
	defer ludp.Close()

	uport, err := strconv.Atoi(u.Port())
	require.Nil(t, err)
	// this is not quite a proper test because there's no dns resolution
	inst := edgeproto.AppInst{
		Uri: "127.0.0.1",
		MappedPorts: []dme.AppPort{
			dme.AppPort{
				Proto:      dme.LProto_LProtoTCP,
				PublicPort: 44444,
			},
			dme.AppPort{
				Proto:      dme.LProto_LProtoUDP,
				PublicPort: 44444,
			},
			dme.AppPort{
				Proto:      dme.LProto_LProtoHTTP,
				PublicPort: int32(uport),
			},
		},
	}
	status := AppInstGetStatus(&inst, timeout)
	require.Equal(t, "ok", status)

	// negative tests
	inst = edgeproto.AppInst{
		Uri: "127.0.0.1",
		MappedPorts: []dme.AppPort{
			dme.AppPort{
				Proto:      dme.LProto_LProtoTCP,
				PublicPort: 64444,
			},
		},
	}
	status = AppInstGetStatus(&inst, timeout)
	fmt.Println(status)
	require.NotEqual(t, "ok", status)
}
