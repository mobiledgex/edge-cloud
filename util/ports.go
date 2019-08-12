package util

import (
	"fmt"
	"strconv"
	"strings"
)

type PortSpec struct {
	Proto string
	Port  string
	EndPort string // mfw XXX ? why two type and parse rtns for AppPort? (3 actually kube.go is another)
}

func ParsePorts(accessPorts string) ([]PortSpec, error) {
	var baseport int64
	ports := []PortSpec{}
	pstrs := strings.Split(accessPorts, ",")
	for _, pstr := range pstrs {
		pp := strings.Split(pstr, ":")
		// within each pp[1], we may have a hypenated range of ports ex: udp:M-N inclusive
		if strings.Contains(pp[1], "-") {
			var err error
			portrange := strings.Split(pp[1], "-")
			baseport, err = strconv.ParseInt(portrange[0], 10, 32)
			if (err != nil) {
				return nil, fmt.Errorf("unable to convert port range base value")
			}
			endport, err := strconv.ParseInt(portrange[1], 10, 32)
			if (err != nil) {
				return nil, fmt.Errorf("unable to convert port range endpoint value")
			}
			// I _think_ all ports orginate from the intial ParseAppPorts in objs.go
			// and as such should have already passed the syntax checks...
			if (baseport < 1 || baseport > 65535) ||
				(endport < 1 || endport > 65535) {
				return nil, fmt.Errorf("Range ports out of range")
			}
			portSpec := PortSpec{
				Proto: pp[0],
				Port:  strconv.FormatInt(baseport, 10),
				EndPort: strconv.FormatInt(endport, 10),
			}
			ports = append(ports, portSpec)
		} else {
			if len(pp) != 2 {
				return nil, fmt.Errorf("invalid AccessPorts format %s", pstr)
			}
			port, err := strconv.ParseInt(pp[1], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("Failed to convert port %s to integer: %s", pp[1], err)
			}
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("Port %s out of range", pp[1])
			}
			portSpec := PortSpec{
				Proto: pp[0],
				Port:  strconv.FormatInt(port, 10),
				EndPort: strconv.FormatInt(0, 32),
			}
			ports = append(ports, portSpec)
		}
	}
	return ports, nil
}
