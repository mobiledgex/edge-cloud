package dmecommon

import (
	"time"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

// Filled in by DME. Added to EdgeEventStatCall to update stats
type GpsLocationInfo struct {
	GpsLocation      *dme.Loc
	SessionCookieKey CookieKey
	Timestamp        dme.Timestamp
	Carrier          string
}

type GpsLocationStats struct {
	Stats     []*GpsLocationInfo
	StartTime time.Time
}

// Distances from Device to AppInst in km
var GpsLocationDists = []int{
	1,
	2,
	5,
	10,
	50,
	100,
}

var GpsLocationQuadrants = []string{
	"Northeast",
	"Southeast",
	"Southwest",
	"Northwest",
}

func NewGpsLocationStats() *GpsLocationStats {
	g := new(GpsLocationStats)
	g.Stats = make([]*GpsLocationInfo, 0)
	g.StartTime = time.Now()
	return g
}
