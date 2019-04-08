package orm

import (
	"net"
	"net/http"
	"testing"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var Success = true
var Fail = false

func TestController(t *testing.T) {
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

	// run dummy controller - this always returns success
	// to all APIs directed to it, and does not actually
	// create or delete objects. We are mocking it out
	// so we can test rbac permissions.
	dc := grpc.NewServer()
	ctrlAddr := "127.0.0.1:9998"
	lis, err := net.Listen("tcp", ctrlAddr)
	require.Nil(t, err)
	ds := testutil.RegisterDummyServer(dc)
	go func() {
		dc.Serve(lis)
	}()
	defer dc.Stop()

	// wait till mc is ready
	err = server.WaitUntilReady()
	require.Nil(t, err, "server online")

	mcClient := &ormclient.Client{}

	// login as super user
	token, err := mcClient.DoLogin(uri, DefaultSuperuser, DefaultSuperpass)
	require.Nil(t, err, "login as superuser")

	// test controller api
	ctrls, status, err := mcClient.ShowController(uri, token)
	require.Nil(t, err, "show controllers")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(ctrls))
	ctrl := ormapi.Controller{
		Region:  "USA",
		Address: ctrlAddr,
	}
	// create controller
	status, err = mcClient.CreateController(uri, token, &ctrl)
	require.Nil(t, err, "create controller")
	require.Equal(t, http.StatusOK, status)
	ctrls, status, err = mcClient.ShowController(uri, token)
	require.Nil(t, err, "show controllers")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(ctrls))
	require.Equal(t, ctrl.Region, ctrls[0].Region)
	require.Equal(t, ctrl.Address, ctrls[0].Address)

	// create admin
	admin, tokenAd := testCreateUser(t, mcClient, uri, "admin1")
	testAddUserRole(t, mcClient, uri, token, "", "AdminManager", admin.Name, Success)

	// create a developers
	org1 := "org1"
	org2 := "org2"
	_, _, tokenDev := testCreateUserOrg(t, mcClient, uri, "dev", "developer", org1)
	_, _, tokenDev2 := testCreateUserOrg(t, mcClient, uri, "dev2", "developer", org2)
	dev3, tokenDev3 := testCreateUser(t, mcClient, uri, "dev3")
	dev4, tokenDev4 := testCreateUser(t, mcClient, uri, "dev4")
	// create an operator
	org3 := "org3"
	org4 := "org4"
	_, _, tokenOper := testCreateUserOrg(t, mcClient, uri, "oper", "operator", org3)
	_, _, tokenOper2 := testCreateUserOrg(t, mcClient, uri, "oper2", "operator", org4)
	oper3, tokenOper3 := testCreateUser(t, mcClient, uri, "oper3")
	oper4, tokenOper4 := testCreateUser(t, mcClient, uri, "oper4")

	// number of fake objects internally sent back by dummy server
	icnt := 3

	// number of dummy objects we add of each type and org
	dcnt := 3
	addDummyObjs(ds, org1, dcnt)
	addDummyObjs(ds, org2, dcnt)
	addDummyObjs(ds, org3, dcnt)
	addDummyObjs(ds, org4, dcnt)

	// number of objects total of each type (sum of above)
	count := icnt + 4*dcnt

	// additional users don't have access to orgs yet
	badPermTestApp(t, mcClient, uri, tokenDev3, ctrl.Region, org1)
	badPermTestShowApp(t, mcClient, uri, tokenDev3, ctrl.Region, org1)

	badPermTestAppInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1)
	badPermTestShowAppInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1)

	badPermTestClusterInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1)
	badPermTestShowClusterInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1)

	badPermTestCloudlet(t, mcClient, uri, tokenOper3, ctrl.Region, org1)

	// add new users to orgs
	testAddUserRole(t, mcClient, uri, tokenDev, org1, "DeveloperContributor", dev3.Name, Success)
	testAddUserRole(t, mcClient, uri, tokenDev, org1, "DeveloperViewer", dev4.Name, Success)
	testAddUserRole(t, mcClient, uri, tokenOper, org3, "OperatorContributor", oper3.Name, Success)
	testAddUserRole(t, mcClient, uri, tokenOper, org3, "OperatorViewer", oper4.Name, Success)
	// make sure dev/ops without user perms can't add new users
	user5, _ := testCreateUser(t, mcClient, uri, "user5")
	testAddUserRole(t, mcClient, uri, tokenDev3, org1, "DeveloperViewer", user5.Name, Fail)
	testAddUserRole(t, mcClient, uri, tokenDev4, org1, "DeveloperViewer", user5.Name, Fail)
	testAddUserRole(t, mcClient, uri, tokenOper3, org3, "OperatorViewer", user5.Name, Fail)
	testAddUserRole(t, mcClient, uri, tokenOper4, org3, "OperatorViewer", user5.Name, Fail)

	// make sure developer and operator cannot see or modify controllers
	ctrlNew := ormapi.Controller{
		Region:  "Bad",
		Address: "bad.mobiledgex.net",
	}
	status, err = mcClient.CreateController(uri, tokenDev, &ctrlNew)
	require.Equal(t, http.StatusForbidden, status)
	status, err = mcClient.CreateController(uri, tokenOper, &ctrlNew)
	require.Equal(t, http.StatusForbidden, status)
	ctrls, status, err = mcClient.ShowController(uri, tokenDev)
	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, 0, len(ctrls))
	ctrls, status, err = mcClient.ShowController(uri, tokenOper)
	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, 0, len(ctrls))

	// admin can do everything
	goodPermTestFlavor(t, mcClient, uri, tokenAd, ctrl.Region, "", icnt)
	goodPermTestClusterFlavor(t, mcClient, uri, tokenAd, ctrl.Region, "", icnt)
	goodPermTestCloudlet(t, mcClient, uri, tokenAd, ctrl.Region, org3, count)
	goodPermTestCloudlet(t, mcClient, uri, tokenAd, ctrl.Region, org4, count)
	goodPermTestApp(t, mcClient, uri, tokenAd, ctrl.Region, org1, dcnt)
	goodPermTestApp(t, mcClient, uri, tokenAd, ctrl.Region, org2, dcnt)
	goodPermTestAppInst(t, mcClient, uri, tokenAd, ctrl.Region, org1, dcnt)
	goodPermTestAppInst(t, mcClient, uri, tokenAd, ctrl.Region, org2, dcnt)
	goodPermTestClusterInst(t, mcClient, uri, tokenAd, ctrl.Region, org1, dcnt)
	goodPermTestClusterInst(t, mcClient, uri, tokenAd, ctrl.Region, org2, dcnt)

	// flavors, clusterflavors, and cloudlets are special - can be seen by all
	goodPermTestShowFlavor(t, mcClient, uri, tokenDev, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenDev2, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenDev3, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenDev4, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenOper, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenOper2, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenOper3, ctrl.Region, "", icnt)
	goodPermTestShowFlavor(t, mcClient, uri, tokenOper4, ctrl.Region, "", icnt)

	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenDev, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenDev2, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenDev3, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenDev4, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenOper, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenOper2, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenOper3, ctrl.Region, "", icnt)
	goodPermTestShowClusterFlavor(t, mcClient, uri, tokenOper4, ctrl.Region, "", icnt)

	goodPermTestShowCloudlet(t, mcClient, uri, tokenDev, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenDev2, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenDev3, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenDev4, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenOper, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenOper2, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenOper3, ctrl.Region, "", count)
	goodPermTestShowCloudlet(t, mcClient, uri, tokenOper4, ctrl.Region, "", count)
	// However, flavors and clusterflavors cannot be modified by non-admins
	badPermTestFlavor(t, mcClient, uri, tokenDev, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenDev2, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenDev3, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenDev4, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenOper, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenOper2, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenOper3, ctrl.Region, "")
	badPermTestFlavor(t, mcClient, uri, tokenOper4, ctrl.Region, "")

	badPermTestClusterFlavor(t, mcClient, uri, tokenDev, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenDev2, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenDev3, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenDev4, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenOper, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenOper2, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenOper3, ctrl.Region, "")
	badPermTestClusterFlavor(t, mcClient, uri, tokenOper4, ctrl.Region, "")

	// make sure operator cannot create apps, appinsts, clusters, etc
	badPermTestApp(t, mcClient, uri, tokenOper, ctrl.Region, org1)
	badPermTestShowApp(t, mcClient, uri, tokenOper, ctrl.Region, org1)

	badPermTestAppInst(t, mcClient, uri, tokenOper, ctrl.Region, org1)
	badPermTestShowAppInst(t, mcClient, uri, tokenOper, ctrl.Region, org1)

	badPermTestClusterInst(t, mcClient, uri, tokenOper, ctrl.Region, org1)
	badPermTestShowClusterInst(t, mcClient, uri, tokenOper, ctrl.Region, org1)

	badPermTestApp(t, mcClient, uri, tokenOper2, ctrl.Region, org1)
	badPermTestShowApp(t, mcClient, uri, tokenOper2, ctrl.Region, org1)

	badPermTestAppInst(t, mcClient, uri, tokenOper2, ctrl.Region, org1)
	badPermTestShowAppInst(t, mcClient, uri, tokenOper2, ctrl.Region, org1)

	badPermTestClusterInst(t, mcClient, uri, tokenOper2, ctrl.Region, org1)
	badPermTestShowClusterInst(t, mcClient, uri, tokenOper2, ctrl.Region, org1)

	// make sure developer cannot create cloudlet (but they can see all of them)
	badPermTestCloudlet(t, mcClient, uri, tokenDev, ctrl.Region, org3)
	badPermTestCloudlet(t, mcClient, uri, tokenDev2, ctrl.Region, org3)

	// test operators can modify their own objs but not each other's
	badPermTestCloudlet(t, mcClient, uri, tokenOper, ctrl.Region, org4)
	badPermTestCloudlet(t, mcClient, uri, tokenOper2, ctrl.Region, org3)

	// test developers can modify their own objs but not each other's
	permTestApp(t, mcClient, uri, tokenDev, tokenDev2, ctrl.Region,
		org1, org2, dcnt)
	permTestAppInst(t, mcClient, uri, tokenDev, tokenDev2, ctrl.Region,
		org1, org2, dcnt)
	permTestClusterInst(t, mcClient, uri, tokenDev, tokenDev2, ctrl.Region,
		org1, org2, dcnt)
	// test users with different roles
	goodPermTestApp(t, mcClient, uri, tokenDev3, ctrl.Region, org1, dcnt)
	goodPermTestAppInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1, dcnt)
	goodPermTestClusterInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1, dcnt)
	// test users with different roles
	goodPermTestCloudlet(t, mcClient, uri, tokenOper3, ctrl.Region, org3, count)
	goodPermTestClusterInst(t, mcClient, uri, tokenDev, ctrl.Region, org1, dcnt)
	badPermTestClusterInst(t, mcClient, uri, tokenDev2, ctrl.Region, org1)

	// remove users from roles, test that they can't modify anything anymore
	testRemoveUserRole(t, mcClient, uri, tokenDev, org1, "DeveloperContributor", dev3.Name, Success)
	badPermTestApp(t, mcClient, uri, tokenDev3, ctrl.Region, org1)
	badPermTestAppInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1)
	badPermTestClusterInst(t, mcClient, uri, tokenDev3, ctrl.Region, org1)
	testRemoveUserRole(t, mcClient, uri, tokenOper, org3, "OperatorContributor", oper3.Name, Success)
	badPermTestCloudlet(t, mcClient, uri, tokenOper3, ctrl.Region, org3)

	// delete controller
	status, err = mcClient.DeleteController(uri, token, &ctrl)
	require.Nil(t, err, "delete controller")
	require.Equal(t, http.StatusOK, status)
	ctrls, status, err = mcClient.ShowController(uri, token)
	require.Nil(t, err, "show controllers")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(ctrls))
}

func testCreateUser(t *testing.T, mcClient *ormclient.Client, uri, name string) (*ormapi.User, string) {
	user := ormapi.User{
		Name:     name,
		Email:    name + "@gmail.com",
		Passhash: name + "-password",
	}
	status, err := mcClient.CreateUser(uri, &user)
	require.Nil(t, err, "create user ", name)
	require.Equal(t, http.StatusOK, status)
	// login
	token, err := mcClient.DoLogin(uri, user.Name, user.Passhash)
	require.Nil(t, err, "login as ", name)
	return &user, token
}

func testCreateOrg(t *testing.T, mcClient *ormclient.Client, uri, token, orgType, orgName string) *ormapi.Organization {
	// create org
	org := ormapi.Organization{
		Type:    orgType,
		Name:    orgName,
		Address: orgName,
		Phone:   "123-123-1234",
	}
	status, err := mcClient.CreateOrg(uri, token, &org)
	require.Nil(t, err, "create org ", orgName)
	require.Equal(t, http.StatusOK, status)
	return &org
}

func testCreateUserOrg(t *testing.T, mcClient *ormclient.Client, uri, name, orgType, orgName string) (*ormapi.User, *ormapi.Organization, string) {
	user, token := testCreateUser(t, mcClient, uri, name)
	org := testCreateOrg(t, mcClient, uri, token, orgType, orgName)
	return user, org, token
}

func testAddUserRole(t *testing.T, mcClient *ormclient.Client, uri, token, org, role, username string, success bool) {
	roleArg := ormapi.Role{
		Username: username,
		Org:      org,
		Role:     role,
	}
	status, err := mcClient.AddUserRole(uri, token, &roleArg)
	if success {
		require.Nil(t, err, "add user role")
		require.Equal(t, http.StatusOK, status)
	} else {
		require.Equal(t, http.StatusForbidden, status)
	}
}

func testRemoveUserRole(t *testing.T, mcClient *ormclient.Client, uri, token, org, role, username string, success bool) {
	roleArg := ormapi.Role{
		Username: username,
		Org:      org,
		Role:     role,
	}
	status, err := mcClient.RemoveUserRole(uri, token, &roleArg)
	require.Nil(t, err, "remove user role")
	require.Equal(t, http.StatusOK, status)
	if success {
	} else {
		require.Equal(t, http.StatusForbidden, status)
	}
}

func setClusterInstDev(dev string, insts []edgeproto.ClusterInst) {
	for ii, _ := range insts {
		insts[ii].Key.Developer = dev
	}
}

func addDummyObjs(d *testutil.DummyServer, org string, num int) {
	for ii := 0; ii < num; ii++ {
		app := edgeproto.App{}
		app.Key.DeveloperKey.Name = org
		d.Apps = append(d.Apps, app)

		appinst := edgeproto.AppInst{}
		appinst.Key.AppKey.DeveloperKey.Name = org
		d.AppInsts = append(d.AppInsts, appinst)

		cinst := edgeproto.ClusterInst{}
		cinst.Key.Developer = org
		d.ClusterInsts = append(d.ClusterInsts, cinst)

		cloudlet := edgeproto.Cloudlet{}
		cloudlet.Key.OperatorKey.Name = org
		d.Cloudlets = append(d.Cloudlets, cloudlet)
	}
}
