package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var accessPorts = []string{
	"tcp:123",                   // 0
	"tcp:123:tls",               // 1
	"tcp:123:nginx",             // 2
	"tcp:0",                     // 3
	"tcp:65536",                 // 4
	"tcp:-4",                    // 5
	"udp:0",                     // 6
	"udp:65536",                 // 7
	"udp:-3",                    // 8
	"udp:20",                    // 9
	"tcp:10-19,udp:20-22",       // 10
	"udp:20-22,udp:23-25:nginx", // 11
	"tcp:1000-1999",             // 12
	"tcp:1000-2000",             // 13
	"udp:1000-1999",             // 14
	"udp:1000-2000",             // 15
	"udp:10000-19999",           // 16
	"udp:10000-20000",           // 17
	"accessports",               // 18
	"http:80",                   // 19
	"udp:20-22,udp:23-25:nginx:maxpktsize=1600", // 20
	"udp:23-25:maxpktsize=1",                    // 21
	"udp:26:maxpktsize=50000",                   // 22
	"tcp:20:maxpktsize=50000",                   // 23
}

func TestParsePorts(t *testing.T) {
	for i, v := range accessPorts {
		ports, err := ParsePorts(v)
		switch i {
		case 0:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, 1, len(ports), "One port")
		case 1:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, true, ports[0].Tls, "tls enabled port")
		case 2:
			require.NotNil(t, err, "no nginx for tcp")
		case 3:
			require.NotNil(t, err, "Port 1-65535 not allowed")
		case 4:
			require.NotNil(t, err, "Ports outside 1-65535 not allowed")
		case 5:
			require.NotNil(t, err, "Ports outside 1-65535 not allowed")
		case 6:
			require.NotNil(t, err, "Ports outside 1-65535 not allowed")
		case 7:
			require.NotNil(t, err, "Ports outside 1-65535 not allowed")
		case 8:
			require.NotNil(t, err, "Ports outside 1-65535 not allowed")
		case 9:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, 1, len(ports), "One udp port")
		case 10:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, "tcp", ports[0].Proto, "tcp protocol")
			require.Equal(t, "udp", ports[1].Proto, "udp protocol")
		case 11:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, false, ports[0].Nginx, "no nginx for ports 20-22")
			require.Equal(t, true, ports[1].Nginx, "nginx for ports 23-25")
		case 12:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, 1, len(ports), "1000 tcp ports")
			require.Equal(t, "1000", ports[0].Port, "port range start")
			require.Equal(t, "1999", ports[0].EndPort, "port range end")
		case 13:
			require.NotNil(t, err, "Not allowed more than 1000 tcp ports")
		case 14:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, false, ports[0].Nginx, "no nginx for under 1000 udp ports unless specified")
		case 15:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, true, ports[0].Nginx, "automatically switch to nginx if over 1000 udp ports")
		case 16:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, true, ports[0].Nginx, "automatically switch to nginx if over 1000 udp ports")
		case 17:
			require.NotNil(t, err, "Not allowed more than 10000 udp ports")
		case 18:
			require.NotNil(t, err, "Incorrect accessports format")
		case 19:
			require.NotNil(t, err, "Only tcp and udp are recognized as valid protocols")
		case 20:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, int64(0), ports[0].MaxPktSize, "valid pkt size can be set for UDP")
			require.Equal(t, int64(1600), ports[1].MaxPktSize, "valid pkt size can be set for UDP")
		case 21:
			require.NotNil(t, err, "Incorrect maxpktsize value")
		case 22:
			require.Nil(t, err, "valid accessPorts input")
			require.Equal(t, int64(50000), ports[0].MaxPktSize, "valid pkt size can be set for UDP")
		case 23:
			require.NotNil(t, err, "maxpktsize not valid for tcp")

		}

	}
}
