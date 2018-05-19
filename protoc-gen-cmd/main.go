package main

import "github.com/mobiledgex/edge-cloud/gensupport"

func main() {
	plugin := GenCmd{}
	gensupport.RunMain("gencmd", ".cmd.go", &plugin, &plugin.support)
}
