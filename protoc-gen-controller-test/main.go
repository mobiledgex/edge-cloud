package main

import "github.com/mobiledgex/edge-cloud/gensupport"

func main() {
	plugin := ControllerTest{}
	gensupport.RunMain("main", ".auto_test.go", &plugin, &plugin.support)
}
