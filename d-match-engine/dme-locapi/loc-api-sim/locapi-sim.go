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
	"strings"
	"time"

	"github.com/AsGz/geo/georeverse"
	dmecommon "github.com/mobiledgex/edge-cloud/d-match-engine/dme-common"
	dmelocapi "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi"
	locutil "github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi/util"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

var locations map[string]dme.Loc

//var countries map[string]string

var (
	locport     = flag.Int("port", 8080, "listen port")
	indexpath   = "/"
	verpath     = "/verifyLocation"
	showlocpath = "/showLocations"
	updatepath  = "/updateLocation"

	locdbfile   = flag.String("file", "/var/tmp/locapisim.yml", "file of IP to location mappings")
	geofile     = flag.String("geo", "", "geocoder db file")
	homeCountry = flag.String("country", "US", "Home Country")
)

var locdbLastUpdate = int64(0)
var countryReverser *georeverse.CountryReverser

func printUsage() {
	fmt.Println("\nUsage: \nlocsim [options]\n\noptions:")
	flag.PrintDefaults()
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("doing showIndex")
	rc := "/verifyLocation -- verifies the location of an token vs lat and long\n"
	rc += "/updateLocation -- adds or replaces an IP->location entry\n"
	rc += "/showLocations -- shows current locations\n"
	w.Write([]byte(rc))
}

func getCountryCode(long, lat float64) string {
	cc := countryReverser.GetCountryCode(long, lat)
	fmt.Printf("Got country code %s for lat %f long %f\n", cc, lat, long)
	return cc
}

func checkForLocUpdate() {
	file, err := os.Stat(*locdbfile)
	if err != nil {
		fmt.Printf("Error in getting locdb file time %v", err)
		locations = make(map[string]dme.Loc)
		return
	}

	modifiedtime := file.ModTime().Unix()
	if modifiedtime > locdbLastUpdate {
		fmt.Printf("need to refresh locations from file\n")
		readLocationFile()
		locdbLastUpdate = time.Now().Unix()
	}
}

func showLocations(w http.ResponseWriter, r *http.Request) {
	log.Printf("doing showLocations %+v\n", locations)
	checkForLocUpdate()
	b, err := json.Marshal(locations)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	w.Write(b)

}

func updateLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("doing updateLocation")
	checkForLocUpdate()

	reqb, err := ioutil.ReadAll(r.Body)
	log.Printf("updateLocation body: %v request %+v\n", string(reqb), r)

	var req dmelocapi.LocationRequestMessage

	err = json.Unmarshal(reqb, &req)
	if err != nil {
		log.Printf("json unmarshall error: %v\n", err)
		http.Error(w, err.Error(), 500)
		return
	}
	if req.Lat == 0 || req.Long == 0 {
		log.Printf("missing field in request:  %+v\n", req)
		http.Error(w, "improperly formatted request", 400)
		return
	}

	// if no ip address provided, take it from the source
	if req.Ipaddress == "" {
		log.Printf("need to take address from %s", r.RemoteAddr)
		remotess := strings.Split(r.RemoteAddr, ":")
		if len(remotess) == 2 {
			req.Ipaddress = remotess[0]
		} else {
			log.Printf("unable to get remote IP %s", r.RemoteAddr)
			http.Error(w, "unable to get remote IP", 400)
			return
		}
	}

	log.Printf("updateLocation addr: %s lat: %f long: %f\n", req.Ipaddress, req.Lat, req.Long)
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
	checkForLocUpdate()
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
		log.Printf("missing field in request %+v\n", req)
		http.Error(w, "missing field", 400)
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
		foundLoc, err := findLocForIP(ip)
		if err != nil {
			resp.MatchingDegree = fmt.Sprintf("%d", dmecommon.LocationUnknown)
		} else {

			reqLoc := dme.Loc{Lat: req.Lat, Long: req.Long}
			log.Printf("find distance between: %+v and %+v\n", reqLoc, foundLoc)
			d := dmecommon.DistanceBetween(reqLoc, foundLoc)
			resp.MatchingDegree = fmt.Sprintf("%d", (dmecommon.GetLocationResultForDistance(d)))

			if *geofile != "" {
				reqCC := countryReverser.GetCountryCode(req.Long, req.Lat)
				realCC := countryReverser.GetCountryCode(foundLoc.Long, foundLoc.Lat)
				log.Printf("country codes req: [%s] real [%s] home [%s]\n", reqCC, realCC, *homeCountry)
				ccMismatch := false
				if reqCC != realCC && reqCC != "" && realCC != "" {
					ccMismatch = true
					log.Printf("mismatched country code\n")
				}
				if reqCC != *homeCountry {
					// roaming
					if ccMismatch {
						resp.MatchingDegree = fmt.Sprintf("%d", dmecommon.LocationRoamingCountryMismatch)
					} else {
						resp.MatchingDegree = fmt.Sprintf("%d", dmecommon.LocationRoamingCountryMatch)
					}
				} else {
					//non roaming
					if ccMismatch {
						resp.MatchingDegree = fmt.Sprintf("%d", dmecommon.LocationMismatchOtherCountry)
					}
				}
			}
			log.Printf("calculated distance: %v km\n match degree: %s", int(d), resp.MatchingDegree)
		}
	}

	respb, _ := json.Marshal(resp)
	log.Printf("Sending response: %v", string(respb))
	w.Write([]byte(string(respb)))
}

func findLocForIP(ipaddr string) (dme.Loc, error) {
	log.Printf("Searching for ip %v\n", ipaddr)

	loc, ok := locations[ipaddr]
	if ok {
		return loc, nil
	}
	log.Printf("unable to find IP %s\n", ipaddr)
	return dme.Loc{}, errors.New("unable to find IP ")
}

func readLocationFile() {
	locations = make(map[string]dme.Loc)
	err := util.ReadYamlFile(*locdbfile, &locations, "", false)
	if err != nil {
		log.Printf("unable to read yaml location file %v\n", err)
	}

}

func run() {
	http.HandleFunc(indexpath, showIndex)
	http.HandleFunc(verpath, verifyLocation)
	http.HandleFunc(updatepath, updateLocation)
	http.HandleFunc(showlocpath, showLocations)

	portstr := fmt.Sprintf(":%d", *locport)

	if *geofile != "" {
		var err error
		countryReverser, err = georeverse.NewCountryReverser(*geofile)
		if err != nil {
			panic(err)
		}
	}
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
	run()
}
