package dmelocapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/mobiledgex/edge-cloud/log"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi/util"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type LocationResponseMessage struct {
	LocationResult uint32 `json:"locresult"`
	Error          string `json:"error"`
}

//format of the HTTP request body.  Token is used for validation of location, but
//IP address is still present to allow locations to be updated for the simulator
type LocationRequestMessage struct {
	Lat       float64       `json:"latitude" yaml:"lat"`
	Long      float64       `json:"longitude" yaml:"long"`
	Token     util.GDDTToken `json:"token" yaml:"token"`
	Ipaddress string        `json:"ipaddr" yaml:"ipaddr"`
}

//REST API client for the GDDT implementation of Location verification API
func CallGDDTLocationVerifyAPI(locVerUrl string, lat, long float64, token string) dmecommon.LocationResult {

	var lrm LocationRequestMessage
	lrm.Lat = lat
	lrm.Long = long
	lrm.Token = util.GDDTToken(token)

	b, _ := json.Marshal(lrm)
	body := bytes.NewBufferString(string(b))
	resp, err := http.Post(locVerUrl, "application/json", body)

	if err != nil {
		log.WarnLog("Error in POST to GDDT Loc service error", "error", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}
	defer resp.Body.Close()

	log.DebugLog(log.DebugLevelLocapi, "Received response", "statusCode:", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:
		log.DebugLog(log.DebugLevelLocapi, "200OK received")

	//treat 401 or 403 as a token issue.  Handling with GDDT to be confirmed
	case http.StatusForbidden:
		fallthrough
	case http.StatusUnauthorized:
		log.WarnLog("returning Match_Engine_Loc_Verify_LOC_ERROR_INVALID_TOKEN", "received code", resp.StatusCode)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_INVALID_TOKEN}
	default:
		log.WarnLog("returning Match_Engine_Loc_Verify_LOC_ERROR_OTHER", "received code", resp.StatusCode)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}

	respBytes, resperr := ioutil.ReadAll(resp.Body)
	if resperr != nil {
		log.WarnLog("Error read response body", "resperr", resperr)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}
	var lrmResp LocationResponseMessage

	err = json.Unmarshal(respBytes, &lrmResp)
	if err != nil {
		log.WarnLog("Error unmarshall response", "respByes", respBytes, "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}

	log.DebugLog(log.DebugLevelLocapi, "unmarshalled location response", "locationResult:", lrmResp.LocationResult)
	if lrmResp.Error != "" {
		log.WarnLog("Error received in token response", "err", lrmResp.Error)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}

	rc := dmecommon.GetDistanceAndStatusForLocationResult(lrmResp.LocationResult)
	log.DebugLog(log.DebugLevelLocapi, "Returning result", "Location Result", rc)

	return rc
}
