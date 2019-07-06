// Validation functions for validating data received
// from an external source - user input, or network data

package util

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// If new valid characters are added here, be sure to update
// the Sanitize functions below to replace the new characters.
var nameMatch = regexp.MustCompile("^[0-9a-zA-Z][-_0-9a-zA-Z .&,!]*$")
var k8sMatch = regexp.MustCompile("^[0-9a-zA-Z][-0-9a-zA-Z.]*$")
var emailMatch = regexp.MustCompile(`(.+)@(.+)\.(.+)`)

// region names are used in Vault approle names, which are very
// restrictive in what characters they allow.
var regionMatch = regexp.MustCompile(`^\w+$`)

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

func ValidEmail(email string) bool {
	return emailMatch.MatchString(email)
}

func ValidRegion(name string) bool {
	return regionMatch.MatchString(name)
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
// a DNS name. Valid chars are only 0-9, a-z, and '-'.
func DNSSanitize(name string) string {
	r := strings.NewReplacer(
		"_", "-",
		" ", "",
		"&", "",
		",", "",
		".", "",
		"!", "")
	return strings.ToLower(r.Replace(name))
}

func K8SSanitize(name string) string {
	r := strings.NewReplacer(
		"_", "-",
		" ", "",
		"&", "",
		",", "",
		"!", "")
	return strings.ToLower(r.Replace(name))
}

// Gitlab groups can only contain letters, digits, _ . -
// cannot start with '-' or end in '.', '.git' or '.atom'
// This combines the rules for both name and path.
func GitlabGroupSanitize(name string) string {
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimSuffix(name, ".")
	if strings.HasSuffix(name, ".git") {
		name = name[:len(name)-4] + "-git"
	}
	if strings.HasSuffix(name, ".atom") {
		name = name[:len(name)-5] + "-atom"
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) ||
			r == '_' || r == '.' || r == '-' {
			return r
		}
		return '-'
	}, name)
}

func ValidOrgName(name string) error {
	re := regexp.MustCompile("^[a-zA-Z0-9_\\-.]*$")
	if !re.MatchString(name) {
		return fmt.Errorf("Name can only contain letters, digits, _ . -")
	}
	if !ValidLDAPName(name) {
		return fmt.Errorf("invalid characters in Name")
	}
	if name != strings.ToLower(name) {
		return fmt.Errorf("Name should have all lowercase characters")
	}
	if strings.Contains(name, "::") {
		return fmt.Errorf("Name cannot contain ::")
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("Name cannot start with '.'")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("Name cannot start with '-'")
	}
	if strings.HasSuffix(name, ".") {
		return fmt.Errorf("Name cannot end with '.'")
	}
	if strings.HasSuffix(name, ".git") {
		return fmt.Errorf("Name cannot end with '.git'")
	}
	if strings.HasSuffix(name, ".atom") {
		return fmt.Errorf("Name cannot end with '.atom'")
	}
	if strings.HasSuffix(name, "-cache") {
		return fmt.Errorf("Name cannot end with '-cache'")
	}
	return nil
}

// IsLatitudeValid checks that the latitude is within accepted ranges
func IsLatitudeValid(latitude float64) bool {
	return (latitude >= -90) && (latitude <= 90)
}

// IsLongitudeValid checks that the longitude is within accepted ranges
func IsLongitudeValid(longitude float64) bool {
	return (longitude >= -180) && (longitude <= 180)
}

func ImagePathParse(imagepath string) (*url.URL, error) {
	// url.Parse requires the scheme but won't error if
	// it's not present.
	if !strings.Contains(imagepath, "://") {
		imagepath = "https://" + imagepath
	}
	return url.Parse(imagepath)
}
