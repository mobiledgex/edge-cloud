package dmelocapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

func CallLocationVerifyAPI(locVerUrl string, lat, long float64, ipaddr string) dme.Match_Engine_Loc_Verify_GPS_Location_Status {

	var lrm LocationRequestMessage
	lrm.Lat = lat
	lrm.Long = long
	lrm.Ipaddress = ipaddr
	b, _ := json.Marshal(lrm)
	body := bytes.NewBufferString(string(b))
	resp, err := http.Post(locVerUrl+"/verifyLocation", "application/json", body)

	if err != nil {
		fmt.Printf("Error in POST to loc service%v\n", err)
		return dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	}
	defer resp.Body.Close()
	fmt.Printf("response Status: %v\n", resp.StatusCode)
	respBytes, err1 := ioutil.ReadAll(resp.Body)

	fmt.Printf("response Status: %v Body: %v\n", resp.StatusCode, string(respBytes))

	if err1 != nil {
		fmt.Printf("Error read response body%v\n", err1)
		return dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	}
	var lrmResp LocationResponseMessage

	err = json.Unmarshal(respBytes, &lrmResp)
	if err != nil {
		fmt.Printf("Error unmarshall response%v\n", err)
		return dme.Match_Engine_Loc_Verify_LOC_UNKNOWN
	}
	fmt.Printf("returning %v lrmresp: %v\n", dme.Match_Engine_Loc_Verify_GPS_Location_Status(lrmResp.LocationResult), lrmResp)

	return dme.Match_Engine_Loc_Verify_GPS_Location_Status(lrmResp.LocationResult)
}
