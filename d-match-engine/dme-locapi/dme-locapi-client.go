package dmelocapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mobiledgex/edge-cloud/log"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
)

type LocationResponseMessage struct {
	LocationResult int32 `json:"locresult"`
}

//format of the HTTP request body
type LocationRequestMessage struct {
	Lat       float64 `json:"lat" yaml:"lat"`
	Long      float64 `json:"long" yaml:"long"`
	Altitude  float64 `json:"altitude" yaml:"altitude"`
	Ipaddress string  `json:"ipaddr" yaml:"ipaddr"`
}

//REST API client for the GDDT implementation of Location verification API
func CallGDDTLocationVerifyAPI(locVerUrl string, lat, long float64, ipaddr string) dme.Match_Engine_Loc_Verify_GPS_Location_Status {

	var lrm LocationRequestMessage
	lrm.Lat = lat
	lrm.Long = long
	lrm.Ipaddress = ipaddr
	b, _ := json.Marshal(lrm)
	body := bytes.NewBufferString(string(b))
	resp, err := http.Post(locVerUrl+"/verifyLocation", "application/json", body)

	if err != nil {
		log.WarnLog("Error in POST to loc service error: %v\n", err)
		return dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	}
	defer resp.Body.Close()
	respBytes, err1 := ioutil.ReadAll(resp.Body)

	log.DebugLog(log.DebugLevelDmelocapi, "Received response", "statusCode:", resp.StatusCode)

	if err1 != nil {
		log.WarnLog("Error read response body:", err1)
		return dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	}
	var lrmResp LocationResponseMessage

	err = json.Unmarshal(respBytes, &lrmResp)
	if err != nil {
		fmt.Printf("Error unmarshall response%v\n", err)
		return dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	}

	log.DebugLog(log.DebugLevelDmelocapi, "unmarshalled location response", "locationResult:", lrmResp.LocationResult)
	return dme.Match_Engine_Loc_Verify_GPS_Location_Status(lrmResp.LocationResult)
}
