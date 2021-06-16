package util

import (
	"fmt"
	"strconv"
	"strings"
)

var maxTcpPorts int = 1000
var maxUdpPorts int = 10000
var maxEnvoyUdpPorts int = 1000
var MaxK8sUdpPorts int = 1000

type PortSpec struct {
	Proto   string
	Port    string
	EndPort string // mfw XXX ? why two type and parse rtns for AppPort? (3 actually kube.go is another)
	Tls     bool
	Nginx   bool
}

func ParsePorts(accessPorts string) ([]PortSpec, error) {
	var baseport int64
	var endport int64
	var err error

	tcpPortCount := 0
	udpPortCount := 0

	ports := []PortSpec{}
	pstrs := strings.Split(accessPorts, ",")

	for _, pstr := range pstrs {
		pp := strings.Split(pstr, ":")
		if len(pp) < 2 {
			return nil, fmt.Errorf("invalid AccessPorts format '%s'", pstr)
		}
		annotations := make(map[string]string)
		for _, kv := range pp[2:] {
			if kv == "" {
				return nil, fmt.Errorf("invalid AccessPorts annotation %s for port %s, expected format is either key or key=val", kv, pp[1])
			}
			keyval := strings.Split(kv, "=")
			if len(keyval) == 1 {
				// boolean annotation
				annotations[kv] = "true"
			} else if len(keyval) == 2 {
				annotations[keyval[0]] = keyval[1]
			} else {
				return nil, fmt.Errorf("invalid AccessPorts annotation %s for port %s, expected format is either key or key=val", kv, pp[1])
			}
		}
		// within each pp[1], we may have a hypenated range of ports ex: udp:M-N inclusive
		portrange := strings.Split(pp[1], "-")
		// len of portrange is 2 if a range 1 if simple port value
		// in either case, baseport is the first elem of portrange
		baseport, err = strconv.ParseInt(portrange[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("unable to convert port range base value")
		}
		if len(portrange) == 2 {
			endport, err = strconv.ParseInt(portrange[1], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("unable to convert port range base value")
			}
		} else {
			endport = baseport
		}

		if (baseport < 1 || baseport > 65535) ||
			(endport < 1 || endport > 65535) {
			return nil, fmt.Errorf("App ports out of range")
		}
		if endport < baseport {
			// after some debate, error on this potential typo len(portrange)
			return nil, fmt.Errorf("App ports out of range")
		}

		if baseport == endport {
			// ex: tcp:5000-5000 or just portrange len = 1
			endport = 0
		}

		proto := strings.ToLower(pp[0])
		if proto != "tcp" && proto != "udp" {
			return nil, fmt.Errorf("Unsupported protocol: %s", pp[0])
		}

		portCount := 1
		if endport != 0 {
			portCount = int(endport-baseport) + 1
		}
		if proto == "tcp" {
			tcpPortCount = tcpPortCount + portCount
		} else { // udp
			udpPortCount = udpPortCount + portCount
		}

		portSpec := PortSpec{
			Proto:   proto,
			Port:    strconv.FormatInt(baseport, 10),
			EndPort: strconv.FormatInt(endport, 10),
		}
		for key, val := range annotations {
			switch key {
			case "tls":
				if portSpec.Proto != "tcp" {
					return nil, fmt.Errorf("Invalid protocol %s, not available for tls support", portSpec.Proto)
				}
				portSpec.Tls = true
			case "nginx":
				if portSpec.Proto != "udp" {
					return nil, fmt.Errorf("Invalid annotation \"nginx\" for %s ports", portSpec.Proto)
				}
				portSpec.Nginx = true
			default:
				return nil, fmt.Errorf("unrecognized annotation %s for port %s", key+"="+val, pp[1])
			}
		}
		ports = append(ports, portSpec)
	}
	if tcpPortCount > maxTcpPorts {
		return nil, fmt.Errorf("Not allowed to specify more than %d tcp ports", maxTcpPorts)
	}
	if udpPortCount > maxUdpPorts {
		return nil, fmt.Errorf("Not allowed to specify more than %d udp ports", maxUdpPorts)
	}
	if udpPortCount > maxEnvoyUdpPorts {
		for i, _ := range ports {
			if ports[i].Proto == "udp" {
				ports[i].Nginx = true
			}
		}
	}

	return ports, nil
}
