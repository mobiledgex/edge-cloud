// Validation functions for validating data received
// from an external source - user input, or network data

package util

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
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

func ValidObjName(name string) error {
	re := regexp.MustCompile("^[a-zA-Z0-9_\\-.]*$")
	if !re.MatchString(name) {
		return fmt.Errorf("name can only contain letters, digits, _ . -")
	}
	if err := ValidLDAPName(name); err != nil {
		return err
	}
	return nil
}

// IsLatitudeValid checks that the latitude is within accepted ranges
func IsLatitudeValid(latitude float64) bool {
	if latitude == 0 {
		return false
	}
	return (latitude >= -90) && (latitude <= 90)
}

// IsLongitudeValid checks that the longitude is within accepted ranges
func IsLongitudeValid(longitude float64) bool {
	if longitude == 0 {
		return false
	}
	return (longitude >= -180) && (longitude <= 180)
}

func ValidateImagePath(imagePath string) error {
	urlInfo := strings.Split(imagePath, "#")
	if len(urlInfo) != 2 {
		return fmt.Errorf("md5 checksum of image is required. Please append checksum to imagepath: \"<url>#md5:checksum\"")
	}
	cSum := strings.Split(urlInfo[1], ":")
	if len(cSum) != 2 {
		return fmt.Errorf("incorrect checksum format, valid format: \"<url>#md5:checksum\"")
	}
	if cSum[0] != "md5" {
		return fmt.Errorf("only md5 checksum is supported")
	}
	if len(cSum[1]) < 32 {
		return fmt.Errorf("md5 checksum must be at least 32 characters")
	}
	_, err := hex.DecodeString(cSum[1])
	if err != nil {
		return fmt.Errorf("invalid md5 checksum")
	}
	return nil
}

func ImagePathParse(imagepath string) (*url.URL, error) {
	// url.Parse requires the scheme but won't error if
	// it's not present.
	if !strings.Contains(imagepath, "://") {
		imagepath = "https://" + imagepath
	}
	return url.Parse(imagepath)
}

func VersionParse(version string) (*time.Time, error) {
	// 2nd Jan 2016
	ref_layout := "2006-01-02"
	vers, err := time.Parse(ref_layout, version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %v", err)
	}
	return &vers, nil
}
