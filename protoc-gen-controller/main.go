package main

import "github.com/mobiledgex/edge-cloud/gensupport"

func main() {
	plugin := ControllerGen{}
	gensupport.RunMain("main", ".auto.go", &plugin, &plugin.support)
}
