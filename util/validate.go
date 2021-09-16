// Validation functions for validating data received
// from an external source - user input, or network data

package util

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// If new valid characters are added here, be sure to update
// the Sanitize functions below to replace the new characters.
var nameMatch = regexp.MustCompile("^[0-9a-zA-Z][-_0-9a-zA-Z .&,!]*$")
var k8sMatch = regexp.MustCompile("^[0-9a-zA-Z][-0-9a-zA-Z.]*$")
var emailMatch = regexp.MustCompile(`(.+)@(.+)\.(.+)`)
var dockerNameMatch = regexp.MustCompile(`^[0-9a-zA-Z][a-zA-Z0-9_.-]+$`)

const maxHostnameLength = 63

// region names are used in Vault approle names, which are very
// restrictive in what characters they allow.
var regionMatch = regexp.MustCompile(`^\w+$`)

// walk through the map of names and values and validate the values
// return an error about which key had invalid value
func ValidateNames(names map[string]string) error {
	if names == nil {
		return nil
	}
	for k, v := range names {
		if v != "" && !ValidName(v) {
			return fmt.Errorf("invalid %s", k)
		}
	}
	return nil
}

func ValidName(name string) bool {
	return nameMatch.MatchString(name)
}

func ValidKubernetesName(name string) bool {
	return k8sMatch.MatchString(name)
}

func ValidDockerName(name string) bool {
	return dockerNameMatch.MatchString(name)
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
// a DNS name. Valid chars are only 0-9, a-z, and '-' and cannot end in '-'
func DNSSanitize(name string) string {
	r := strings.NewReplacer(
		"_", "-",
		" ", "",
		"&", "",
		",", "",
		".", "",
		"!", "")
	rval := strings.ToLower(r.Replace(name))
	return strings.TrimRight(rval, "-")
}

// HostnameSanitize makes a valid hostname, for which the rules
// are the same as DNSSanitize, but it cannot end in '-' and cannot
// be > 63 digits
func HostnameSanitize(name string) string {
	r := DNSSanitize(name)
	if len(r) > maxHostnameLength {
		r = r[:maxHostnameLength]
	}
	return strings.TrimRight(r, "-")
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

// alphanumeric plus -_. first char must be alpha, <= 255 chars.
func HeatSanitize(name string) string {
	r := strings.NewReplacer(
		" ", "",
		"&", "",
		",", "",
		"!", "")
	str := r.Replace(name)
	if str == "" {
		return str
	}
	if !unicode.IsLetter(rune(str[0])) {
		// first character must be alpha
		str = "a" + str
	}
	if len(str) > 255 {
		str = str[:254]
	}
	return str
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
	return (latitude >= -90) && (latitude <= 90)
}

// IsLongitudeValid checks that the longitude is within accepted ranges
func IsLongitudeValid(longitude float64) bool {
	return (longitude >= -180) && (longitude <= 180)
}

func ValidateImagePath(imagePath string) error {
	url, err := ImagePathParse(imagePath)
	if err != nil {
		return fmt.Errorf("invalid image path: %v", err)
	}
	ext := filepath.Ext(url.Path)
	if ext == "" {
		return fmt.Errorf("missing filename from image path")
	}

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
	_, err = hex.DecodeString(cSum[1])
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

func ContainerVersionParse(version string) (*time.Time, error) {
	// 2nd Jan 2016
	ref_layout := "2006-01-02"
	vers, err := time.Parse(ref_layout, version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse container version: %v", err)
	}
	return &vers, nil
}

func ValidateImageVersion(imgVersion string) error {
	re := regexp.MustCompile("^[0-9a-zA-Z][-0-9a-zA-Z._]*$")
	if !re.MatchString(imgVersion) {
		return fmt.Errorf("ImageVersion can only contain letters, digits, -, ., _")
	}
	return nil
}

func ValidK8SContainerName(name string) error {
	parts := strings.Split(name, "/")
	if len(parts) == 1 {
		if !ValidKubernetesName(name) {
			return fmt.Errorf("Invalid kubernetes container name")
		}
	} else if len(parts) == 2 {
		if !ValidKubernetesName(parts[0]) {
			return fmt.Errorf("Invalid kubernetes pod name")
		}
		if !ValidKubernetesName(parts[1]) {
			return fmt.Errorf("Invalid kubernetes container name")
		}
	} else if len(parts) == 3 {
		if !ValidKubernetesName(parts[0]) {
			return fmt.Errorf("Invalid kubernetes namespace name")
		}
		if !ValidKubernetesName(parts[1]) {
			return fmt.Errorf("Invalid kubernetes pod name")
		}
		if !ValidKubernetesName(parts[2]) {
			return fmt.Errorf("Invalid kubernetes container name")
		}
	} else {
		return fmt.Errorf("Invalid kubernetes container name, should be of format '<namespace>/<PodName>/<ContainerName>'")
	}
	return nil
}
