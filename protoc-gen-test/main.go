package main

import "github.com/mobiledgex/edge-cloud/gensupport"

func main() {
	testcud := TestCud{}
	gensupport.RunMain("testutil", "_testutil.go", &testcud, &testcud.support)
}
