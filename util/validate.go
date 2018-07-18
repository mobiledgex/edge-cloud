// Validation functions for validating data received
// from an external source - user input, or network data

package util

import (
	"net"
	"regexp"
	"strings"
)

var AppDNSRoot = "mobiledgex.net"

// If new valid characters are added here, be sure to update
// the Sanitize functions below to replace the new characters.
var nameMatch = regexp.MustCompile("^[0-9a-zA-Z][-_0-9a-zA-Z .&,!]*$")
var k8sMatch = regexp.MustCompile("^[0-9a-zA-Z][-0-9a-zA-Z.]*$")

func ValidName(name string) bool {
	return nameMatch.MatchString(name)
}

func ValidKubernetesName(name string) bool {
	return k8sMatch.MatchString(name)
}

func ValidIp(ip []byte) bool {
	if len(ip) != net.IPv4len && len(ip) != net.IPv6len {
		return false
	}
	return true
}

// DockerSanitize sanitizes the name string (which is assumed to be a
// ValidName) to make it usable as a docker image name
// (no spaces and special chars other than - and . are allowed)
func DockerSanitize(name string) string {
	r := strings.NewReplacer(
		" ", "",
		"&", "-",
		",", "",
		"!", ".")
	return r.Replace(name)
}

// DNSSanitize santizies the name string to make it usable in
// a DNS name. Valid chars are only 0-9, a-z (both cases), and '-'.
func DNSSanitize(name string) string {
	r := strings.NewReplacer(
		"_", "-",
		" ", "",
		"&", "",
		",", "",
		".", "",
		"!", "")
	return r.Replace(name)
}

func K8SSanitize(name string) string {
	r := strings.NewReplacer(
		"_", "-",
		" ", "",
		"&", "",
		",", "",
		"!", "")
	return r.Replace(name)
}
