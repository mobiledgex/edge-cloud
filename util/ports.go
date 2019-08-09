package util

import (
	"fmt"
	"strconv"
	"strings"
)

type PortSpec struct {
	Proto string
	Port  string
}

func ParsePorts(accessPorts string) ([]PortSpec, error) {
	ports := []PortSpec{}
	pstrs := strings.Split(accessPorts, ",")
	for _, pstr := range pstrs {
		pp := strings.Split(pstr, ":")
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
		}
		ports = append(ports, portSpec)
	}
	return ports, nil
}
