package orm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/vault"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
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

	err = server.WaitUntilReady()
	require.Nil(t, err, "server online")

	// login as super user
	token, err := ormclient.DoLogin(uri, DefaultSuperuser, DefaultSuperpass)
	require.Nil(t, err, "login as superuser")

	super, status, err := showCurrentUser(uri, token)
	require.Nil(t, err, "show super")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, DefaultSuperuser, super.Name, "super user name")
	require.Equal(t, "", super.Passhash, "empty pass hash")
	require.Equal(t, "", super.Salt, "empty salt")
	require.Equal(t, 0, super.Iter, "empty iter")

	roleAssignments, status, err := ormclient.ShowRoleAssignment(uri, token)
	require.Nil(t, err, "show roles")
	require.Equal(t, http.StatusOK, status, "show role status")
	require.Equal(t, 1, len(roleAssignments), "num role assignments")
	require.Equal(t, RoleAdminManager, roleAssignments[0].Role)
	require.Equal(t, super.Name, roleAssignments[0].Username)

	// show users - only super user at this point
	users, status, err := ormclient.ShowUsers(uri, token, "")
	require.Equal(t, http.StatusOK, status, "show user status")
	require.Equal(t, 1, len(users))
	require.Equal(t, DefaultSuperuser, users[0].Name, "super user name")
	require.Equal(t, "", users[0].Passhash, "empty pass hash")
	require.Equal(t, "", users[0].Salt, "empty salt")
	require.Equal(t, 0, users[0].Iter, "empty iter")

	policies, status, err := showRolePerms(uri, token)
	require.Nil(t, err, "show role perms err")
	require.Equal(t, http.StatusOK, status, "show role perms status")
	require.Equal(t, 82, len(policies), "number of role perms")
	roles, status, err := showRoles(uri, token)
	require.Nil(t, err, "show roles err")
	require.Equal(t, http.StatusOK, status, "show roles status")
	require.Equal(t, 9, len(roles), "number of roles")

	// create new user1
	user1 := ormapi.User{
		Name:     "MisterX",
		Email:    "misterx@gmail.com",
		Passhash: "misterx-password",
	}
	status, err = ormclient.CreateUser(uri, &user1)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user1
	tokenMisterX, err := ormclient.DoLogin(uri, user1.Name, user1.Passhash)
	require.Nil(t, err, "login as mister X")
	// create an Organization
	org1 := ormapi.Organization{
		Type:    "developer",
		Name:    "DevX",
		Address: "123 X Way",
		Phone:   "123-123-1234",
	}
	status, err = ormclient.CreateOrg(uri, tokenMisterX, &org1)
	require.Nil(t, err, "create org")
	require.Equal(t, http.StatusOK, status, "create org status")

	// create new user2
	user2 := ormapi.User{
		Name:     "MisterY",
		Email:    "mistery@gmail.com",
		Passhash: "mistery-password",
	}
	status, err = ormclient.CreateUser(uri, &user2)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user2
	tokenMisterY, err := ormclient.DoLogin(uri, user2.Name, user2.Passhash)
	require.Nil(t, err, "login as mister Y")
	// create an Organization
	org2 := ormapi.Organization{
		Type:    "developer",
		Name:    "DevY",
		Address: "123 Y Way",
		Phone:   "123-321-1234",
	}
	status, err = ormclient.CreateOrg(uri, tokenMisterY, &org2)
	require.Nil(t, err, "create org")
	require.Equal(t, http.StatusOK, status, "create org status")

	// check org membership as mister x
	orgs, status, err := ormclient.ShowOrgs(uri, tokenMisterX)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(orgs))
	require.Equal(t, org1.Name, orgs[0].Name)
	require.Equal(t, org1.Type, orgs[0].Type)
	// check org membership as mister y
	orgs, status, err = ormclient.ShowOrgs(uri, tokenMisterY)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(orgs))
	require.Equal(t, org2.Name, orgs[0].Name)
	require.Equal(t, org2.Type, orgs[0].Type)
	// super user should be able to show all orgs
	orgs, status, err = ormclient.ShowOrgs(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(orgs))

	// check role assignments as mister x
	roleAssignments, status, err = ormclient.ShowRoleAssignment(uri, tokenMisterX)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(roleAssignments))
	require.Equal(t, user1.Name, roleAssignments[0].Username)
	// check role assignments as mister y
	roleAssignments, status, err = ormclient.ShowRoleAssignment(uri, tokenMisterY)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(roleAssignments))
	require.Equal(t, user2.Name, roleAssignments[0].Username)
	// super user should be able to see all role assignments
	roleAssignments, status, err = ormclient.ShowRoleAssignment(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, len(roleAssignments))

	// show org users as mister x
	users, status, err = ormclient.ShowUsers(uri, tokenMisterX, org1.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
	require.Equal(t, user1.Name, users[0].Name)
	// show org users as mister y
	users, status, err = ormclient.ShowUsers(uri, tokenMisterY, org2.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
	require.Equal(t, user2.Name, users[0].Name)
	// super user can see all users with org ID = 0
	users, status, err = ormclient.ShowUsers(uri, token, "")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, len(users))

	// check that x and y cannot see each other's org users
	users, status, err = ormclient.ShowUsers(uri, tokenMisterX, org2.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	users, status, err = ormclient.ShowUsers(uri, tokenMisterY, org1.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	users, status, err = ormclient.ShowUsers(uri, tokenMisterX, "foobar")
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// check that x and y cannot delete each other's orgs
	status, err = ormclient.DeleteOrg(uri, tokenMisterX, &org2)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	status, err = ormclient.DeleteOrg(uri, tokenMisterY, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// delete orgs
	status, err = ormclient.DeleteOrg(uri, tokenMisterX, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = ormclient.DeleteOrg(uri, tokenMisterY, &org2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	// delete users
	status, err = ormclient.DeleteUser(uri, tokenMisterX, &user1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = ormclient.DeleteUser(uri, tokenMisterY, &user2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// check orgs are gone
	orgs, status, err = ormclient.ShowOrgs(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(orgs))
	// check users are gone
	users, status, err = ormclient.ShowUsers(uri, token, "")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
}

func showCurrentUser(uri, token string) (*ormapi.User, int, error) {
	user := ormapi.User{}
	status, err := ormclient.PostJson(uri+"/auth/user/current", token, nil, &user)
	return &user, status, err
}

func showRolePerms(uri, token string) ([]ormapi.RolePerm, int, error) {
	perms := []ormapi.RolePerm{}
	status, err := ormclient.PostJson(uri+"/auth/role/perms/show", token, nil, &perms)
	return perms, status, err
}

func showRoles(uri, token string) ([]string, int, error) {
	roles := []string{}
	status, err := ormclient.PostJson(uri+"/auth/role/show", token, nil, &roles)
	return roles, status, err
}

func waitServerOnline(addr string) error {
	return fmt.Errorf("wait server online failed")
}

// for debug
func dumpTables() {
	users := []ormapi.User{}
	orgs := []ormapi.Organization{}
	db.Find(&users)
	db.Find(&orgs)
	for _, user := range users {
		fmt.Printf("%v\n", user)
	}
	for _, org := range orgs {
		fmt.Printf("%v\n", org)
	}
}
