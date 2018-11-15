package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/mobiledgex/edge-cloud/deploygen"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestDeployGen(t *testing.T) {
	go main()

	time.Sleep(time.Second)
	for ii, app := range testutil.AppData {
		for gen, _ := range deploygen.Generators {
			url := "http://" + *addr + "/" + gen
			fmt.Printf("posting to %s\n", url)
			mf, err := deploygen.SendReq(url, &app)
			require.Nil(t, err, "post app[%d] to %s", ii, gen)
			fmt.Printf("AppData[%d]:\n", ii)
			fmt.Printf("%s\n", mf)

			mf2, err := deploygen.RunGen(gen, &app)
			require.Nil(t, err, "run app[%d] to %s", ii, gen)
			require.Equal(t, mf, mf2, "equal app[%d] gen %s", ii, gen)
		}
	}
}
