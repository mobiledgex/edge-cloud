package orm

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestAllData(t *testing.T) {
	log.SetDebugLevel(log.DebugLevelApi)
	addr := "127.0.0.1:9999"
	uri := "http://" + addr + "/api/v1"

	config := ServerConfig{
		ServAddr:  addr,
		SqlAddr:   "127.0.0.1:5445",
		RunLocal:  true,
		InitLocal: true,
		IgnoreEnv: true,
	}
	server, err := RunServer(&config)
	require.Nil(t, err, "run server")
	defer server.Stop()

	Jwks.Init("addr", "mcorm", "roleID", "secretID")
	Jwks.Meta.CurrentVersion = 1
	Jwks.Keys[1] = &vault.JWK{
		Secret:  "12345",
		Refresh: "1s",
	}

	// wait till mc is ready
	err = server.WaitUntilReady()
	require.Nil(t, err, "server online")

	// admin user is already created
	adminUser := ormapi.User{
		Name:     DefaultSuperuser,
		Passhash: DefaultSuperpass,
	}
	users := []ormapi.User{
		ormapi.User{
			Name:     "dev1",
			Email:    "dev1@email.com",
			Passhash: "dev1password",
		},
		ormapi.User{
			Name:     "dev2",
			Email:    "dev2@email.com",
			Passhash: "dev2password",
		},
	}
	admindata := `
controllers:
- region: usa
  address: test.usa.mc.mobiledgex.com:1234
- region: germany
  address: test.germany.mc.mobiledgex.com:1234
- region: south korea
  address: test.southkorea.mc.mobiledgex.com:1234
`
	dev1data := `
orgs:
- name: devorg
  type: developer
  address: somewhere
  phone: somenumber
  adminusername: dev1

roles:
- org: devorg
  username: dev1
  role: DeveloperManager
- org: devorg
  username: dev2
  role: DeveloperViewer
`

	// create users
	for _, user := range users {
		status, err := ormclient.CreateUser(uri, &user)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)
	}

	testData(t, uri, &adminUser, admindata)
	testData(t, uri, &users[0], dev1data)
}

func testData(t *testing.T, uri string, user *ormapi.User, yamldata string) {
	data := &ormapi.AllData{}
	err := yaml.Unmarshal([]byte(yamldata), data)
	require.Nil(t, err, "unmarshal yaml")

	// login as specified user
	token, err := ormclient.DoLogin(uri, user.Name, user.Passhash)
	require.Nil(t, err, "login for %s", user.Name)

	// run create
	status, err := ormclient.CreateData(uri, token, data, func(res *ormapi.Result) {
		fmt.Println(res)
	})
	require.Nil(t, err, "create admin data")
	require.Equal(t, http.StatusOK, status)

	// compare optoins
	copts := []cmp.Option{
		cmpopts.IgnoreTypes(time.Time{}),
		cloudcommon.IgnoreAdminRole,
	}

	// run show and compare
	showData, status, err := ormclient.ShowData(uri, token)
	if !cmp.Equal(data, showData, copts...) {
		mismatch := cmp.Diff(data, showData, copts...)
		require.True(t, false, "show data mismatch\n%s", mismatch)
	}

	// run delete
	status, err = ormclient.DeleteData(uri, token, data, func(res *ormapi.Result) {
		fmt.Println(res)
	})
	require.Nil(t, err, "delete data")
	require.Equal(t, http.StatusOK, status)

	// show and compare empty data
	emptyData := ormapi.AllData{}
	showData, status, err = ormclient.ShowData(uri, token)
	if !cmp.Equal(&emptyData, showData, copts...) {
		mismatch := cmp.Diff(&emptyData, showData, copts...)
		require.True(t, false, "show empty data mismatch\n%s", mismatch)
	}
}
