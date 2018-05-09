package main

import "github.com/mobiledgex/edge-cloud/util"

func main() {
	testcud := TestCud{}
	util.RunMain("testutil", "_testutil.go", &testcud)
}
