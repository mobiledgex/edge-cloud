package dmelocapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/mobiledgex/edge-cloud/log"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	"github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi/util"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type LocationResponseMessage struct {
	MatchingDegree string `json:"matchingDegree"`
	Error          string `json:"error"`
}

//format of the HTTP request body.  Token is used for validation of location, but
//IP address is still present to allow locations to be updated for the simulator
type LocationRequestMessage struct {
	Lat       float64       `json:"latitude" yaml:"lat"`
	Long      float64       `json:"longitude" yaml:"long"`
	Token     util.TDGToken `json:"token" yaml:"token"`
	Ipaddress string        `json:"ipaddr,omitempty" yaml:"ipaddr"`
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// CallTDGLocationVerifyAPI REST API client for the TDG implementation of Location verification API
func CallTDGLocationVerifyAPI(locVerUrl string, lat, long float64, token string) dmecommon.LocationResult {

	var lrm LocationRequestMessage
	lrm.Lat = lat
	lrm.Long = long
	lrm.Token = util.TDGToken(token)

	b, err := json.Marshal(lrm)
	if err != nil {
		log.WarnLog("error in json mashal of request", "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}

	body := bytes.NewBufferString(string(b))
	req, err := http.NewRequest("POST", locVerUrl, body)

	if err != nil {
		log.WarnLog("error in http.NewRequest", "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}
	req.Header.Add("Content-Type", "application/json")
	username := os.Getenv("LOCAPI_USER")
	password := os.Getenv("LOCAPI_PASSWD")

	if username != "" {
		log.DebugLog(log.DebugLevelLocapi, "adding auth header", "username", username)
		req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	} else {
		log.DebugLog(log.DebugLevelLocapi, "no auth credentials")
	}
	client := &http.Client{}
	log.DebugLog(log.DebugLevelLocapi, "sending to api gw", "body:", body)

	resp, err := client.Do(req)

	if err != nil {
		log.WarnLog("Error in POST to TDG Loc service error", "error", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}
	defer resp.Body.Close()

	log.DebugLog(log.DebugLevelLocapi, "Received response", "statusCode:", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:
		log.DebugLog(log.DebugLevelLocapi, "200OK received")

	//treat 401 or 403 as a token issue.  Handling with TDG to be confirmed
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

	//resp = string(respBytes)
	err = json.Unmarshal(respBytes, &lrmResp)
	if err != nil {
		log.WarnLog("Error unmarshall response", "respBytes", respBytes, "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}

	log.DebugLog(log.DebugLevelLocapi, "unmarshalled location response", "match degree:", lrmResp.MatchingDegree)
	if lrmResp.Error != "" {
		log.WarnLog("Error received in token response", "err", lrmResp.Error)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}
	l, err := strconv.ParseInt(lrmResp.MatchingDegree, 10, 32)
	if err != nil {
		log.WarnLog("Error in LocationResult", "LocationResult", lrmResp.MatchingDegree, "err", err)
		return dmecommon.LocationResult{DistanceRange: -1, MatchEngineLocStatus: dme.Match_Engine_Loc_Verify_LOC_ERROR_OTHER}
	}
	rc := dmecommon.GetDistanceAndStatusForLocationResult(uint32(l))
	fmt.Printf("6666 %v %v", l, rc)

	log.DebugLog(log.DebugLevelLocapi, "Returning result", "Location Result", rc)

	return rc
}
