package util

import (
	"regexp"
	"strings"
)

// User names and Org names are used in LDAP Distinguished Names
// which need to escape a bunch of special chars by converting to hex.
// Ldap special chars are: , \ # + < > ; " = leading/trailing spaces.
// For convenience, we only allow the , special char.
// Org names are also used as "group" names in gitlab.
var ldapNameMatch = regexp.MustCompile("^[-_0-9a-zA-Z .&,!]+$")

func ValidLDAPName(name string) bool {
	if name == "" {
		return false
	}
	if name[0] == ' ' || name[len(name)-1] == ' ' {
		// ldap: no leading or trailing spaces
		return false
	}
	return ldapNameMatch.MatchString(name)
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
