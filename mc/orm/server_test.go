package orm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
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
	token, err := doLogin(uri, DefaultSuperuser, DefaultSuperpass)
	require.Nil(t, err, "login as superuser")

	super, status, err := showCurrentUser(uri, token)
	require.Nil(t, err, "show super")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, int64(1), super.ID)
	require.Equal(t, DefaultSuperuser, super.Name, "super user name")
	require.Equal(t, "", super.Passhash, "empty pass hash")
	require.Equal(t, "", super.Salt, "empty salt")
	require.Equal(t, 0, super.Iter, "empty iter")

	roleAssignments, status, err := showRoleAssignment(uri, token)
	require.Nil(t, err, "show roles")
	require.Equal(t, http.StatusOK, status, "show role status")
	require.Equal(t, 1, len(roleAssignments), "num role assignments")
	require.Equal(t, RoleAdminManager, roleAssignments[0].Role)
	require.Equal(t, super.ID, roleAssignments[0].UserID)

	// show users - only super user at this point
	users, status, err := showUsers(uri, token, "")
	require.Equal(t, http.StatusOK, status, "show user status")
	require.Equal(t, 1, len(users))
	require.Equal(t, DefaultSuperuser, users[0].Name, "super user name")
	require.Equal(t, "", users[0].Passhash, "empty pass hash")
	require.Equal(t, "", users[0].Salt, "empty salt")
	require.Equal(t, 0, users[0].Iter, "empty iter")

	policies, status, err := showRolePerms(uri, token)
	require.Nil(t, err, "show role perms err")
	require.Equal(t, http.StatusOK, status, "show role perms status")
	require.Equal(t, 78, len(policies), "number of role perms")
	roles, status, err := showRoles(uri, token)
	require.Nil(t, err, "show roles err")
	require.Equal(t, http.StatusOK, status, "show roles status")
	require.Equal(t, 18, len(roles), "number of roles")

	// create new user1
	user1 := User{
		Name:     "MisterX",
		Email:    "misterx@gmail.com",
		Passhash: "misterx-password",
	}
	status, err = createUser(uri, &user1)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user1
	tokenMisterX, err := doLogin(uri, user1.Name, user1.Passhash)
	require.Nil(t, err, "login as mister X")
	// create an Organization
	org1 := Organization{
		Type:    "developer",
		Name:    "DevX",
		Address: "123 X Way",
		Phone:   "123-123-1234",
	}
	status, err = createOrg(uri, tokenMisterX, &org1)
	require.Nil(t, err, "create org")
	require.Equal(t, http.StatusOK, status, "create org status")

	// create new user2
	user2 := User{
		Name:     "MisterY",
		Email:    "mistery@gmail.com",
		Passhash: "mistery-password",
	}
	status, err = createUser(uri, &user2)
	require.Nil(t, err, "create user")
	require.Equal(t, http.StatusOK, status, "create user status")
	// login as new user2
	tokenMisterY, err := doLogin(uri, user2.Name, user2.Passhash)
	require.Nil(t, err, "login as mister Y")
	// create an Organization
	org2 := Organization{
		Type:    "developer",
		Name:    "DevY",
		Address: "123 Y Way",
		Phone:   "123-321-1234",
	}
	status, err = createOrg(uri, tokenMisterY, &org2)
	require.Nil(t, err, "create org")
	require.Equal(t, http.StatusOK, status, "create org status")

	// check org membership as mister x
	orgs, status, err := showOrgs(uri, tokenMisterX)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(orgs))
	require.Equal(t, org1.Name, orgs[0].Name)
	require.Equal(t, org1.Type, orgs[0].Type)
	// check org membership as mister y
	orgs, status, err = showOrgs(uri, tokenMisterY)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(orgs))
	require.Equal(t, org2.Name, orgs[0].Name)
	require.Equal(t, org2.Type, orgs[0].Type)
	// super user should be able to show all orgs
	orgs, status, err = showOrgs(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 2, len(orgs))

	// check role assignments as mister x
	roleAssignments, status, err = showRoleAssignment(uri, tokenMisterX)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(roleAssignments))
	require.Equal(t, user1.ID, roleAssignments[0].UserID)
	// check role assignments as mister y
	roleAssignments, status, err = showRoleAssignment(uri, tokenMisterY)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(roleAssignments))
	require.Equal(t, user2.ID, roleAssignments[0].UserID)
	// super user should be able to see all role assignments
	roleAssignments, status, err = showRoleAssignment(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, len(roleAssignments))

	// show org users as mister x
	users, status, err = showUsers(uri, tokenMisterX, org1.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
	require.Equal(t, user1.Name, users[0].Name)
	// show org users as mister y
	users, status, err = showUsers(uri, tokenMisterY, org2.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
	require.Equal(t, user2.Name, users[0].Name)
	// super user can see all users with org ID = 0
	users, status, err = showUsers(uri, token, "")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 3, len(users))

	// check that x and y cannot see each other's org users
	users, status, err = showUsers(uri, tokenMisterX, org2.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	users, status, err = showUsers(uri, tokenMisterY, org1.Name)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	users, status, err = showUsers(uri, tokenMisterX, "foobar")
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// check that x and y cannot delete each other's orgs
	status, err = deleteOrg(uri, tokenMisterX, &org2)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)
	status, err = deleteOrg(uri, tokenMisterY, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusForbidden, status)

	// delete orgs
	status, err = deleteOrg(uri, tokenMisterX, &org1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = deleteOrg(uri, tokenMisterY, &org2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	// delete users
	status, err = deleteUser(uri, tokenMisterX, &user1)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	status, err = deleteUser(uri, tokenMisterY, &user2)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)

	// check orgs are gone
	orgs, status, err = showOrgs(uri, token)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 0, len(orgs))
	// check users are gone
	users, status, err = showUsers(uri, token, "")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, 1, len(users))
}

func doLogin(uri, user, pass string) (string, error) {
	login := UserLogin{
		Username: user,
		Password: pass,
	}
	result := make(map[string]interface{})
	status, err := postJson(uri+"/login", "", &login, &result)
	if err != nil {
		return "", fmt.Errorf("login error, %s", err.Error())
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("login status %d instead of OK(200)", status)
	}
	tokenI, ok := result["token"]
	if !ok {
		return "", fmt.Errorf("login token not found in response")
	}
	token, ok := tokenI.(string)
	if !ok {
		return "", fmt.Errorf("login token not string")
	}
	return token, nil
}

func showCurrentUser(uri, token string) (*User, int, error) {
	user := User{}
	status, err := postJson(uri+"/auth/user/current", token, nil, &user)
	return &user, status, err
}

func showUsers(uri, token, org string) ([]User, int, error) {
	users := []User{}
	in := Organization{Name: org}
	status, err := postJson(uri+"/auth/user/show", token, &in, &users)
	return users, status, err
}

func showOrgs(uri, token string) ([]Organization, int, error) {
	orgs := []Organization{}
	status, err := postJson(uri+"/auth/org/show", token, nil, &orgs)
	return orgs, status, err
}

func showRolePerms(uri, token string) ([]RolePerm, int, error) {
	perms := []RolePerm{}
	status, err := postJson(uri+"/auth/role/perms/show", token, nil, &perms)
	return perms, status, err
}

func showRoles(uri, token string) ([]string, int, error) {
	roles := []string{}
	status, err := postJson(uri+"/auth/role/show", token, nil, &roles)
	return roles, status, err
}

func showRoleAssignment(uri, token string) ([]Role, int, error) {
	roles := []Role{}
	status, err := postJson(uri+"/auth/role/assignment/show", token, nil, &roles)
	return roles, status, err
}

func createUser(uri string, user *User) (int, error) {
	result := ResultID{}
	status, err := postJson(uri+"/usercreate", "", user, &result)
	if err == nil && status == http.StatusOK {
		user.ID = result.ID
	}
	return status, err
}

func createOrg(uri, token string, org *Organization) (int, error) {
	result := ResultName{}
	status, err := postJson(uri+"/auth/org/create", token, org, &result)
	return status, err
}

func deleteOrg(uri, token string, org *Organization) (int, error) {
	return postJson(uri+"/auth/org/delete", token, org, nil)
}

func deleteUser(uri, token string, user *User) (int, error) {
	return postJson(uri+"/auth/user/delete", token, user, nil)
}

func addUserRole(uri, token string, role *Role) (int, error) {
	result := Result{}
	status, err := postJson(uri+"/auth/role/adduser", token, role, &result)
	return status, err
}

func removeUserRole(uri, token string, role *Role) (int, error) {
	result := Result{}
	status, err := postJson(uri+"/auth/role/removeuser", token, role, &result)
	return status, err
}

func postJsonSend(uri, token string, reqData interface{}) (*http.Response, error) {
	var body io.Reader
	if reqData != nil {
		out, err := json.Marshal(reqData)
		if err != nil {
			return nil, fmt.Errorf("post %s marshal req failed, %s", uri, err.Error())
		}
		body = bytes.NewBuffer(out)
	} else {
		body = nil
	}
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, fmt.Errorf("post %s http req failed, %s", uri, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	client := &http.Client{}
	return client.Do(req)
}

func postJson(uri, token string, reqData interface{}, replyData interface{}) (int, error) {
	resp, err := postJsonSend(uri, token, reqData)
	if err != nil {
		return 0, fmt.Errorf("post %s client do failed, %s", uri, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK && replyData != nil {
		err = json.NewDecoder(resp.Body).Decode(replyData)
		if err != nil {
			return 0, fmt.Errorf("post %s decode resp failed, %s", uri, err.Error())
		}
	}
	return resp.StatusCode, nil
}

func postJsonStreamOut(uri, token string, reqData interface{}, replyData interface{}) (int, error) {
	resp, err := postJsonSend(uri, token, reqData)
	if err != nil {
		return 0, fmt.Errorf("post %s client do failed, %s", uri, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK && replyData != nil {
		dec := json.NewDecoder(resp.Body)
		for {
			err := dec.Decode(replyData)
			if err != nil {
				if err == io.EOF {
					break
				}
				return 0, fmt.Errorf("post %s decode resp failed, %s", uri, err.Error())
			}
		}
	}
	return resp.StatusCode, nil
}

func waitServerOnline(addr string) error {
	return fmt.Errorf("wait server online failed")
}

// for debug
func dumpTables() {
	users := []User{}
	orgs := []Organization{}
	db.Find(&users)
	db.Find(&orgs)
	for _, user := range users {
		fmt.Printf("%v\n", user)
	}
	for _, org := range orgs {
		fmt.Printf("%v\n", org)
	}
}
