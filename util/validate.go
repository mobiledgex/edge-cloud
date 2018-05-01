// Validation functions for validating data received
// from an external source - user input, or network data

package util

import (
	"net"
	"regexp"
)

var nameMatch = regexp.MustCompile("^[0-9a-zA-Z][-_0-9a-zA-Z .&]*$")

func ValidName(name string) bool {
	return nameMatch.MatchString(name)
}

func ValidIp(ip []byte) bool {
	if len(ip) != net.IPv4len && len(ip) != net.IPv6len {
		return false
	}
	return true
}
