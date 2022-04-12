// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
