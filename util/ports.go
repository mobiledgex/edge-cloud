package util

import (
	"fmt"
	"strconv"
	"strings"
)

type PortSpec struct {
	Proto   string
	Port    string
	EndPort string // mfw XXX ? why two type and parse rtns for AppPort? (3 actually kube.go is another)
}

func ParsePorts(accessPorts string) ([]PortSpec, error) {
	var baseport int64
	var endport int64
	var err error

	ports := []PortSpec{}
	pstrs := strings.Split(accessPorts, ",")

	for _, pstr := range pstrs {
		pp := strings.Split(pstr, ":")
		if len(pp) != 2 {
			return nil, fmt.Errorf("Invalid AccessPorts format %s", pstr)
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
			return nil, fmt.Errorf("Range ports out of range")
		}
		if endport < baseport {
			// after some debate, error on this potential typo len(portrange)
			return nil, fmt.Errorf("App ports out of range")
		}

		if baseport == endport {
			// ex: tcp:5000-5000 or just portrange len = 1
			endport = 0
		}

		portSpec := PortSpec{
			Proto:   strings.ToLower(pp[0]),
			Port:    strconv.FormatInt(baseport, 10),
			EndPort: strconv.FormatInt(endport, 10),
		}
		ports = append(ports, portSpec)
	}

	return ports, nil
}
