package ormclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mobiledgex/edge-cloud/mc/ormapi"
)

func DoLogin(uri, user, pass string) (string, error) {
	login := ormapi.UserLogin{
		Username: user,
		Password: pass,
	}
	result := make(map[string]interface{})
	status, err := PostJson(uri+"/login", "", &login, &result)
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

func CreateUser(uri string, user *ormapi.User) (int, error) {
	return PostJson(uri+"/usercreate", "", user, nil)
}

func DeleteUser(uri, token string, user *ormapi.User) (int, error) {
	return PostJson(uri+"/auth/user/delete", token, user, nil)
}

func ShowUsers(uri, token, org string) ([]ormapi.User, int, error) {
	users := []ormapi.User{}
	in := ormapi.Organization{Name: org}
	status, err := PostJson(uri+"/auth/user/show", token, &in, &users)
	return users, status, err
}

func CreateController(uri, token string, ctrl *ormapi.Controller) (int, error) {
	return PostJson(uri+"/auth/controller/create", token, ctrl, nil)
}

func DeleteController(uri, token string, ctrl *ormapi.Controller) (int, error) {
	return PostJson(uri+"/auth/controller/delete", token, ctrl, nil)
}

func ShowController(uri, token string) ([]ormapi.Controller, int, error) {
	ctrls := []ormapi.Controller{}
	status, err := PostJson(uri+"/auth/controller/show", token, nil, &ctrls)
	return ctrls, status, err
}

func CreateOrg(uri, token string, org *ormapi.Organization) (int, error) {
	return PostJson(uri+"/auth/org/create", token, org, nil)
}

func DeleteOrg(uri, token string, org *ormapi.Organization) (int, error) {
	return PostJson(uri+"/auth/org/delete", token, org, nil)
}

func ShowOrgs(uri, token string) ([]ormapi.Organization, int, error) {
	orgs := []ormapi.Organization{}
	status, err := PostJson(uri+"/auth/org/show", token, nil, &orgs)
	return orgs, status, err
}

func AddUserRole(uri, token string, role *ormapi.Role) (int, error) {
	return PostJson(uri+"/auth/role/adduser", token, role, nil)
}

func RemoveUserRole(uri, token string, role *ormapi.Role) (int, error) {
	return PostJson(uri+"/auth/role/removeuser", token, role, nil)
}

func ShowRoleAssignment(uri, token string) ([]ormapi.Role, int, error) {
	roles := []ormapi.Role{}
	status, err := PostJson(uri+"/auth/role/assignment/show", token, nil, &roles)
	return roles, status, err
}

func CreateData(uri, token string, data *ormapi.AllData, cb func(res *ormapi.Result)) (int, error) {
	res := ormapi.Result{}
	status, err := PostJsonStreamOut(uri+"/auth/data/create", token, data, &res, func() {
		cb(&res)
	})
	return status, err
}

func DeleteData(uri, token string, data *ormapi.AllData, cb func(res *ormapi.Result)) (int, error) {
	res := ormapi.Result{}
	status, err := PostJsonStreamOut(uri+"/auth/data/delete", token, data, &res, func() {
		cb(&res)
	})
	return status, err
}

func ShowData(uri, token string) (*ormapi.AllData, int, error) {
	data := ormapi.AllData{}
	status, err := PostJson(uri+"/auth/data/show", token, nil, &data)
	return &data, status, err
}

func PostJsonSend(uri, token string, reqData interface{}) (*http.Response, error) {
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

func PostJson(uri, token string, reqData interface{}, replyData interface{}) (int, error) {
	resp, err := PostJsonSend(uri, token, reqData)
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

func PostJsonStreamOut(uri, token string, reqData interface{}, replyData interface{}, replyReady func()) (int, error) {
	resp, err := PostJsonSend(uri, token, reqData)
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
			if replyReady != nil {
				replyReady()
			}
		}
	}
	return resp.StatusCode, nil
}
