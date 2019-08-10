package simulatedlocation

import (
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
)

var fakeLocation = dme.Loc{Latitude: 32.013988, Longitude: -96.598243}

// VerifySimulatedClientLoc uses a fixed position for the "real" location of the client and tests against that
func VerifySimulatedClientLoc(mreq *dme.VerifyLocationRequest, mreply *dme.VerifyLocationReply) error {

	mreply.GpsLocationStatus = dme.VerifyLocationReply_LOC_UNKNOWN
	mreply.GpsLocationAccuracyKm = -1

	distance := dmecommon.DistanceBetween(*mreq.GpsLocation, fakeLocation)

	//here we want the verified range of the distance based on the distance rules
	//e.g. if the actual distance is 50KM we may return a range to say within 100KM
	locresult := dmecommon.GetLocationResultForDistance(distance)
	locval := dmecommon.GetDistanceAndStatusForLocationResult(locresult)

	mreply.GpsLocationStatus = locval.MatchEngineLocStatus
	mreply.GpsLocationAccuracyKm = locval.DistanceRange

	log.DebugLog(log.DebugLevelDmereq, "verified location at",
		"lat", mreq.GpsLocation.Latitude,
		"long", mreq.GpsLocation.Longitude,
		"actual distance", distance,
		"distance range", mreply.GpsLocationAccuracyKm,
		"status", mreply.GpsLocationStatus)

	return nil
}

func GetSimulatedClientLoc(mreq *dme.GetLocationRequest, reply *dme.GetLocationReply) error {
	reply.CarrierName = mreq.CarrierName
	reply.Status = dme.GetLocationReply_LOC_FOUND
	reply.NetworkLocation = &fakeLocation
	return nil
}
