package main

//very simple app which is similar to the example app, except this one supports HTTP GET

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var (
	port        = flag.Int("port", 8080, "listen port")
	indexpath   = "/"
	getdatapath = "/getdata"
)

func showIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("doing showIndex")
	rc := getdatapath + "\n"
	w.Write([]byte(rc))
}
func getData(w http.ResponseWriter, r *http.Request) {
	log.Printf("doing getData %+v\n", r)

	b := []byte("Z")
	numbytes := uint32(0)
	nb := r.URL.Query().Get("numbytes")

	if nb != "" {
		u, err := strconv.ParseUint(nb, 10, 32)
		if err == nil {
			numbytes = uint32(u)
		} else {
			log.Printf("Error in parseUint %v\n", err)
		}
	}
	response := string(bytes.Repeat(b, int(numbytes)))

	//force non chunked response
	w.Header().Set("Content-Length", nb)
	w.Write([]byte(response))
}

func run() {
	http.HandleFunc(indexpath, showIndex)
	http.HandleFunc(getdatapath, getData)

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
