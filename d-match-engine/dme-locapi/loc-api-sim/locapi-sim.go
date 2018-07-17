package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmelocapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	locutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi/util"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var locations map[string]dme.Loc

var (
	locport     = flag.Int("port", 8080, "listen port")
	indexpath   = "/"
	verpath     = "/verifyLocation"
	updatepath  = "/updateLocation"
	showlocpath = "/showLocations"

	locdbfile = flag.String("file", "/var/tmp/locapisim.yml", "file of IP to location mappings")
)

func printUsage() {
	fmt.Println("\nUsage: \nlocsim [options]\n\noptions:")
	flag.PrintDefaults()
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("doing showIndex")
	rc := "/verifyLocation -- verifies the location of an token vs lat and long\n"
	rc += "/updateLocation -- adds or replaces and IP->location entry\n"
	rc += "/showLocations -- shows current locations\n"
	w.Write([]byte(rc))
}

func showLocations(w http.ResponseWriter, r *http.Request) {
	log.Printf("doing showLocations %+v\n", locations)
	b, err := json.Marshal(locations)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write(b)
}

func updateLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("doing updateLocation")

	reqb, err := ioutil.ReadAll(r.Body)
	log.Printf("body: %v\n", string(reqb))

	var req dmelocapi.LocationRequestMessage

	err = json.Unmarshal(reqb, &req)
	if err != nil {
		log.Printf("json unmarshall error: %v\n", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if req.Ipaddress == "" || req.Lat == 0 || req.Long == 0 {
		log.Printf("missing field in request:  %+v\n", req)
		http.Error(w, "improperly formatted request", 400)
		return
	}

	locations[req.Ipaddress] = dme.Loc{Lat: req.Lat, Long: req.Long}

	ymlout, err := yaml.Marshal(locations)
	if err != nil {
		log.Printf("Error in yaml marshal of location db: %v\n", err)
		http.Error(w, err.Error(), 500)
	} else {
		ofile, err := os.OpenFile(*locdbfile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		defer ofile.Close()
		if err != nil {
			log.Fatalf("unable to write to file: %s, err: %v\n", *locdbfile, err)
		}
		fmt.Fprintf(ofile, string(ymlout))
	}

	w.Write([]byte("Location DB Updated OK for " + req.Ipaddress + "\n"))

}

func verifyLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("doing verifyLocation")

	reqb, err := ioutil.ReadAll(r.Body)
	log.Printf("body: %v\n", string(reqb))

	var req dmelocapi.LocationRequestMessage
	var resp dmelocapi.LocationResponseMessage

	err = json.Unmarshal(reqb, &req)
	if err != nil {
		log.Printf("json unmarshall error: %v\n", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if req.Token == "" || req.Lat == 0 || req.Long == 0 {
		log.Println("missing field in request")
		http.Error(w, err.Error(), 400)
	}

	ip, valid, err := locutil.DecodeToken(req.Token)
	if !valid {
		http.Error(w, "Token is not valid or expired", 401)
		return
	} else if err != nil {
		//likely a badly formatted token
		http.Error(w, err.Error(), 400)
		return
	} else {
		err, foundLoc := findLocForIp(ip)
		if err != nil {
			resp.LocationResult = dmecommon.LocationUnknown
		} else {
			reqLoc := dme.Loc{Lat: req.Lat, Long: req.Long}
			log.Printf("find distance between: %+v and %+v\n", reqLoc, foundLoc)
			d := dmecommon.DistanceBetween(reqLoc, foundLoc)
			resp.LocationResult = dmecommon.GetLocationResultForDistance(d)
			log.Printf("calculated distance: %v km\n result: %d", int(d), resp.LocationResult)
		}
	}

	respb, _ := json.Marshal(resp)
	log.Printf("Sending response: %v", string(respb))
	w.Write([]byte(string(respb)))
}

func findLocForIp(ipaddr string) (error, dme.Loc) {
	log.Printf("Searching for ip %v\n", ipaddr)

	loc, ok := locations[ipaddr]
	if ok {
		return nil, loc
	}
	log.Printf("unable to find IP %s\n", ipaddr)
	return errors.New("unable to find IP "), dme.Loc{}
}

func readLocationFile() {
	if _, err := os.Stat(*locdbfile); err != nil {
		fmt.Printf("file " + *locdbfile + " does not exist, will be created\n")
		locations = make(map[string]dme.Loc)
	} else {
		err := util.ReadYamlFile(*locdbfile, &locations, "", false)
		if err != nil {
			log.Fatalf("unable to read yaml location file %v\n", err)
		}
	}
}

func run() {
	http.HandleFunc(indexpath, showIndex)

	http.HandleFunc(verpath, verifyLocation)
	http.HandleFunc(updatepath, updateLocation)
	http.HandleFunc(showlocpath, showLocations)

	portstr := fmt.Sprintf(":%d", *locport)

	log.Printf("Listening on http://127.0.0.1:%d", *locport)
	if err := http.ListenAndServe(portstr, nil); err != nil {
		panic(err)
	}
}

func validateArgs() {
	flag.Parse()
	//nothing to check yet
}

func main() {
	validateArgs()
	readLocationFile()
	run()
}
