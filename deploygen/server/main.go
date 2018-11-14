package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mobiledgex/edge-cloud/deploygen"
)

var addr = flag.String("addr", "127.0.0.1:8000", "listener address")

func main() {
	flag.Parse()
	for gen, _ := range deploygen.Generators {
		http.HandleFunc("/"+gen, handler)
	}
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	app := deploygen.AppSpec{}
	err := json.NewDecoder(r.Body).Decode(&app)
	if err != nil {
		processErr(w, err)
		return
	}
	gen := strings.TrimLeft(r.RequestURI, "/")
	fx, found := deploygen.Generators[gen]
	if !found {
		processErr(w, fmt.Errorf("generator for %s not found", r.RequestURI))
		return
	}
	out, err := fx(&app)
	if err != nil {
		processErr(w, err)
		return
	}
	fmt.Fprint(w, out)
}

func processErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
