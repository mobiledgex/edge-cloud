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
	"testing"

	"github.com/stretchr/testify/require"
)

type StringMap struct {
	from string
	to   string
}

var camelCaseMaps = []StringMap{
	{
		from: "NewMatch_Engine_ApiClient",
		to:   "NewMatchEngineApiClient",
	},
	{
		from: "GpsLocation",
		to:   "GpsLocation",
	},
	{
		from: "_GPSLocation",
		to:   "GpsLocation",
	},
	{
		from: "L_PROTO_UNKNOWN",
		to:   "LProtoUnknown",
	},
	{
		from: "LProto_TCP",
		to:   "LProtoTcp",
	},
	{
		from: "GPS_Location_Accuracy_KM",
		to:   "GpsLocationAccuracyKm",
	},
	{
		from: "fqdn",
		to:   "Fqdn",
	},
	{
		from: "FQDNs",
		to:   "FqdNs",
	},
	{
		from: "FQDNPrefix",
		to:   "FqdnPrefix",
	},
	{
		from: "FQDN",
		to:   "Fqdn",
	},
	{
		from: "CFKey",
		to:   "CfKey",
	},
}

var camelCaseSplits = map[string][]string{
	"":             []string{},
	"testStr":      []string{"test", "Str"},
	"testStrSS":    []string{"test", "Str", "S", "S"},
	"TTtTestStrSS": []string{"T", "Tt", "Test", "Str", "S", "S"},
	"TestStr12":    []string{"Test", "Str12"},
}

func TestCamelCase(t *testing.T) {
	for _, stringMap := range camelCaseMaps {
		require.Equal(t, stringMap.to, CamelCase(stringMap.from))
	}
	for camelCaseStr, camelSplit := range camelCaseSplits {
		out := SplitCamelCase(camelCaseStr)
		require.Equal(t, len(out), len(camelSplit), camelCaseStr)
		for ii, _ := range out {
			require.Equal(t, out[ii], camelSplit[ii], camelCaseStr)
		}
	}
}

type StringMapError struct {
	from string
	to   string
	err  string
}

func TestQuoteArgs(t *testing.T) {
	tests := []StringMapError{{
		from: "hostname;  hostname",
		to:   `"hostname;" "hostname"`,
	}, {
		from: "ls -ltrh",
		to:   `"ls" "-ltrh"`,
	}, {
		from: `"ab","cd"`,
		to:   `"ab,cd"`,
	}, {
		from: ``,
		to:   ``,
	}, {
		from: `echo "newpassword" > /var/etc/password`,
		to:   `"echo" "newpassword" ">" "/var/etc/password"`,
	}, {
		from: `bash -c "ls -ltrh"`,
		to:   `"bash" "-c" "ls -ltrh"`,
	}, {
		from: `bash -c 'ls -ltrh'`,
		to:   `"bash" "-c" "ls -ltrh"`,
	}, {
		from: `bash -c "ls -ltrh`,
		err:  "Unterminated double-quoted string",
	}, {
		from: `bash -c 'ls -ltrh`,
		err:  "Unterminated single-quoted string",
	}, {
		from: `bash -c \`,
		err:  "Unterminated backslash-escape",
	}, {
		to:   `"bash" "-c" "ls -ltrh"`, // see if func is idempotent
		from: `"bash" "-c" "ls -ltrh"`,
	}}
	for _, test := range tests {
		quoted, err := QuoteArgs(test.from)
		if test.err == "" {
			require.Nil(t, err)
			require.Equal(t, test.to, quoted, "convert %s --> %s", test.from, test.to)
		} else {
			require.NotNil(t, err)
			require.Contains(t, err.Error(), test.err, "quote %s", test.to)
		}
	}
}
