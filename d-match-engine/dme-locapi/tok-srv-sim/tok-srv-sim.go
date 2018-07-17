package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mobiledgex/edge-cloud/d-match-engine/dme-locapi/util"
)

var (
	port      = flag.Int("port", 8080, "listen port")
	indexpath = "/"

	getTokenPath        = "/its"
	getExpiredTokenPath = "/itsexpired"
)

func printUsage() {
	fmt.Println("\nUsage: \token-server-sim [options]\n\noptions:")
	flag.PrintDefaults()
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("doing showIndex")
	rc := "/its -- Identity Token Server Get Token\n"
	w.Write([]byte(rc))
}

func getToken(w http.ResponseWriter, r *http.Request) {
	log.Println("doing getToken")

	f := r.URL.Query().Get("followURL")
	remoteAddr := r.RemoteAddr

	//requests using "localhost" may yield the IPv6 equivalent, force it to IPv4
	remoteAddr = strings.Replace(remoteAddr, "[::1]", "127.0.0.1", -1)
	remoteIp := strings.Split(remoteAddr, ":")[0]

	//the encoding of token for now is just a base64 version of the ip address plus some
	//expiry time.  We will decode this within the token server simulator and use the IP to derive
	//a location, or reject if the expiry time is passed
	token64 := ""
	if strings.Contains(r.URL.Path, getExpiredTokenPath) {
		log.Println("getting an expired token")
		//this is to test the case where we have an expired token. Ask for a token which expired 10 seconds ago.
		token64 = util.GenerateToken(remoteIp, -10)
	} else {
		token64 = util.GenerateToken(remoteIp, util.DefaultTokenValidSeconds)
	}
	log.Printf("followurl: %s remoteIp: %s token: %s\n", f, remoteIp, token64)

	http.Redirect(w, r, f+"?dt-id="+token64, 303)
}

func run() {
	http.HandleFunc(indexpath, showIndex)
	http.HandleFunc(getTokenPath, getToken)
	http.HandleFunc(getExpiredTokenPath, getToken)

	portstr := fmt.Sprintf(":%d", *port)

	log.Printf("Listening on http://127.0.0.1:%d", *port)
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
