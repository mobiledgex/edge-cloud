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

func TestQuoteArgs(t *testing.T) {
	tests := []StringMap{{
		from: "hostname;  hostname",
		to:   `"hostname;" "hostname"`,
	}, {
		from: "ls -ltrh",
		to:   `"ls" "-ltrh"`,
	}, {
		from: `echo "newpassword" > /var/etc/password`,
		to:   `"echo" "newpassword" ">" "/var/etc/password"`,
	}, {
		from: `bash -c "ls -ltrh"`,
		to:   `"bash" "-c" "ls -ltrh"`,
	}}
	for _, test := range tests {
		require.Equal(t, test.to, QuoteArgs(test.from), "convert %s --> %s", test.from, test.to)
	}
}
