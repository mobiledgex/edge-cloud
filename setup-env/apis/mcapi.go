package apis

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/mc/ormclient"
	"github.com/mobiledgex/edge-cloud/setup-env/util"
)

func RunMcAPI(api, mcname, apiFile, curUserFile, outputDir string) bool {
	mc := util.GetMC(mcname)
	uri := "http://" + mc.Addr + "/api/v1"
	log.Printf("Using MC %s at %s", mc.Name, uri)

	if strings.HasSuffix(api, "users") {
		return runMcUsersAPI(api, uri, apiFile, curUserFile, outputDir)
	}
	return runMcDataAPI(api, uri, apiFile, curUserFile, outputDir)
}

func runMcUsersAPI(api, uri, apiFile, curUserFile, outputDir string) bool {
	log.Printf("Applying MC users via APIs for %s\n", apiFile)

	rc := true
	if api == "showusers" {
		token, rc := loginCurUser(uri, curUserFile)
		if !rc {
			return false
		}
		users, status, err := ormclient.ShowUsers(uri, token, "")
		checkMcErr("ShowUser", status, err, &rc)
		util.PrintToYamlFile("show-commands.yml", outputDir, users, true)
		return rc
	}

	if apiFile == "" {
		log.Println("Error: Cannot run MC user APIs without API file")
		return false
	}
	users := readUsersFiles(apiFile)

	switch api {
	case "createusers":
		for _, user := range users {
			status, err := ormclient.CreateUser(uri, &user)
			checkMcErr("CreateUser", status, err, &rc)
		}
	case "deleteusers":
		token, ok := loginCurUser(uri, curUserFile)
		if !ok {
			return false
		}
		for _, user := range users {
			status, err := ormclient.DeleteUser(uri, token, &user)
			checkMcErr("DeleteUser", status, err, &rc)
		}
	}
	return rc
}

func runMcDataAPI(api, uri, apiFile, curUserFile, outputDir string) bool {
	log.Printf("Applying MC data via APIs for %s\n", apiFile)

	// Data APIs are all run by a given user.
	// That user is specified in the current user file.
	// We need to log in as that user.
	rc := true
	token, rc := loginCurUser(uri, curUserFile)
	if !rc {
		return false
	}

	if api == "show" {
		showData, status, err := ormclient.ShowData(uri, token)
		checkMcErr("ShowData", status, err, &rc)
		util.PrintToYamlFile("show-commands.yml", outputDir, showData, true)
		return rc
	}

	if apiFile == "" {
		log.Println("Error: Cannot run MC data APIs without API file")
		return false
	}
	data := readMCDataFile(apiFile)
	switch api {
	case "create":
		status, err := ormclient.CreateData(uri, token, data, func(res *ormapi.Result) {
			log.Printf("CreateData: %s\n", res.Message)
		})
		checkMcErr("CreateData", status, err, &rc)
	case "delete":
		status, err := ormclient.DeleteData(uri, token, data, func(res *ormapi.Result) {
			log.Printf("DeleteData: %s\n", res.Message)
		})
		checkMcErr("DeleteData", status, err, &rc)
	}
	return rc
}

func readUsersFiles(file string) []ormapi.User {
	users := []ormapi.User{}
	files := strings.Split(file, ",")
	for _, file := range files {
		fileusers := []ormapi.User{}
		err := util.ReadYamlFile(file, &fileusers, "", false)
		if err != nil {
			if !util.IsYamlOk(err, "mcusers") {
				fmt.Fprintf(os.Stderr, "error in unmarshal for file %s\n", file)
				os.Exit(1)
			}
		}
		users = append(users, fileusers...)
	}
	return users
}

func readMCDataFile(file string) *ormapi.AllData {
	data := ormapi.AllData{}
	err := util.ReadYamlFile(file, &data, "", false)
	if err != nil {
		if !util.IsYamlOk(err, "mcdata") {
			fmt.Fprintf(os.Stderr, "error in unmarshal for file %s\n", file)
			os.Exit(1)
		}
	}
	return &data
}

func loginCurUser(uri, curUserFile string) (string, bool) {
	if curUserFile == "" {
		log.Println("Error: Cannot run MC APIs without current user file")
		return "", false
	}
	users := readUsersFiles(curUserFile)
	if len(users) == 0 {
		log.Printf("no user to run MC create api\n")
		return "", false
	}
	token, err := ormclient.DoLogin(uri, users[0].Name, users[0].Passhash)
	rc := true
	checkMcErr("DoLogin", http.StatusOK, err, &rc)
	return token, rc
}

func checkMcErr(msg string, status int, err error, rc *bool) {
	if err != nil || status != http.StatusOK {
		log.Printf("%s failed %v/%d\n", msg, err, status)
		*rc = false
	}
}
