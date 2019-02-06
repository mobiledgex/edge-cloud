package main

import "github.com/mobiledgex/edge-cloud/gensupport"

func main() {
	plugin := GenMC2{}
	gensupport.RunMain("orm", ".mc2.go", &plugin, &plugin.support)
}
