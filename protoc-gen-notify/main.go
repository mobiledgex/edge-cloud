package main

import "github.com/mobiledgex/edge-cloud/gensupport"

func main() {
	plugin := GenNotify{}
	gensupport.RunMain("notify", ".notify.go", &plugin, &plugin.support)
}
