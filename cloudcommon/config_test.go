package cloudcommon

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/stretchr/testify/require"
)

func TestAppConfig(t *testing.T) {
	_, err := ParseAppConfig(testConfigStr)
	require.Nil(t, err)

	app := &testutil.AppData[0]
	app.Config = testConfigStr
	checkAppConfig(t, app)

	// test request over http
	tsManifest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testConfigStr)
	}))
	defer tsManifest.Close()

	app.Config = tsManifest.URL
	checkAppConfig(t, app)
}

func checkAppConfig(t *testing.T, app *edgeproto.App) {
	str, err := GetAppConfig(app)
	require.Nil(t, err)
	_, err = ParseAppConfig(str)
	require.Nil(t, err)
}

var testConfigStr = `
resources: some-resource
`
