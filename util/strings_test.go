package util

import (
	"github.com/stretchr/testify/require"
	"testing"
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

func TestCamelCase(t *testing.T) {
	for _, stringMap := range camelCaseMaps {
		require.Equal(t, stringMap.to, CamelCase(stringMap.from))
	}
}
