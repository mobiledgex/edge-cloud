package util

import (
	"fmt"
	"regexp"
	"strings"
)

// User names and Org names are used in LDAP Distinguished Names
// which need to escape a bunch of special chars by converting to hex.
// Ldap special chars are: , \ # + < > ; " = leading/trailing spaces.
// For convenience, we only allow the , special char.
// Org names are also used as "group" names in gitlab.
var ldapNameMatch = regexp.MustCompile("^[-_0-9a-zA-Z .&,!]+$")

func ValidLDAPName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if name[0] == ' ' || name[len(name)-1] == ' ' {
		// ldap: no leading or trailing spaces
		return fmt.Errorf("name cannot have leading or trailing spaces")
	}
	if !ldapNameMatch.MatchString(name) {
		return fmt.Errorf("name does not match LDAP required format")
	}
	return nil
}

func EscapeLDAPName(name string) string {
	r := strings.NewReplacer(
		",", "\\2c")
	return r.Replace(name)
}

func UnescapeLDAPName(name string) string {
	r := strings.NewReplacer(
		"\\2c", ",")
	return r.Replace(name)
}
