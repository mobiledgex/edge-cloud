package orm

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
)

const ActionView = "view"
const ActionManage = "manage"

const ResourceControllers = "controllers"
const ResourceUsers = "users"
const ResourceApps = "apps"
const ResourceAppInsts = "appinsts"
const ResourceClusters = "clusters"
const ResourceClusterInsts = "clusterinsts"
const ResourceAppAnalytics = "appanalytics"
const ResourceCloudlets = "cloudlets"
const ResourceCloudletAnalytics = "cloudletanalytics"
const ResourceClusterFlavors = "clusterflavors"
const ResourceFlavors = "flavors"

var DeveloperResources = []string{
	ResourceApps,
	ResourceAppInsts,
	ResourceClusters,
	ResourceClusterInsts,
	ResourceAppAnalytics,
}
var OperatorResources = []string{
	ResourceCloudlets,
	ResourceCloudletAnalytics,
}

// built-in roles
const RoleDeveloperManager = "DeveloperManager"
const RoleDeveloperContributor = "DeveloperContributor"
const RoleDeveloperViewer = "DeveloperViewer"
const RoleOperatorManager = "OperatorManager"
const RoleOperatorContributor = "OperatorContributor"
const RoleOperatorViewer = "OperatorViewer"
const RoleAdminManager = "AdminManager"
const RoleAdminContributor = "AdminContributor"
const RoleAdminViewer = "AdminViewer"

var AdminRoleID int64

func InitRolePerms() error {
	log.DebugLog(log.DebugLevelApi, "init roleperms")

	enforcer.AddPolicy(RoleAdminManager, ResourceControllers, ActionManage)
	enforcer.AddPolicy(RoleAdminManager, ResourceControllers, ActionView)
	enforcer.AddPolicy(RoleAdminManager, ResourceClusterFlavors, ActionManage)
	enforcer.AddPolicy(RoleAdminManager, ResourceClusterFlavors, ActionView)
	enforcer.AddPolicy(RoleAdminManager, ResourceFlavors, ActionManage)
	enforcer.AddPolicy(RoleAdminManager, ResourceFlavors, ActionView)

	enforcer.AddPolicy(RoleDeveloperManager, ResourceUsers, ActionManage)
	enforcer.AddPolicy(RoleDeveloperManager, ResourceUsers, ActionView)
	enforcer.AddPolicy(RoleOperatorManager, ResourceUsers, ActionManage)
	enforcer.AddPolicy(RoleOperatorManager, ResourceUsers, ActionView)
	enforcer.AddPolicy(RoleAdminManager, ResourceUsers, ActionManage)
	enforcer.AddPolicy(RoleAdminManager, ResourceUsers, ActionView)
	for _, str := range DeveloperResources {
		enforcer.AddPolicy(RoleDeveloperManager, str, ActionManage)
		enforcer.AddPolicy(RoleDeveloperManager, str, ActionView)
		enforcer.AddPolicy(RoleDeveloperContributor, str, ActionManage)
		enforcer.AddPolicy(RoleDeveloperContributor, str, ActionView)
		enforcer.AddPolicy(RoleDeveloperViewer, str, ActionView)
		enforcer.AddPolicy(RoleAdminManager, str, ActionManage)
		enforcer.AddPolicy(RoleAdminManager, str, ActionView)
		enforcer.AddPolicy(RoleAdminContributor, str, ActionManage)
		enforcer.AddPolicy(RoleAdminContributor, str, ActionView)
		enforcer.AddPolicy(RoleAdminViewer, str, ActionView)
	}
	for _, str := range OperatorResources {
		enforcer.AddPolicy(RoleOperatorManager, str, ActionManage)
		enforcer.AddPolicy(RoleOperatorManager, str, ActionView)
		enforcer.AddPolicy(RoleOperatorContributor, str, ActionManage)
		enforcer.AddPolicy(RoleOperatorContributor, str, ActionView)
		enforcer.AddPolicy(RoleOperatorViewer, str, ActionView)
		enforcer.AddPolicy(RoleAdminManager, str, ActionManage)
		enforcer.AddPolicy(RoleAdminManager, str, ActionView)
		enforcer.AddPolicy(RoleAdminContributor, str, ActionManage)
		enforcer.AddPolicy(RoleAdminContributor, str, ActionView)
		enforcer.AddPolicy(RoleAdminViewer, str, ActionView)
	}
	return nil
}

func ShowRolePerms(c echo.Context) error {
	_, err := getClaims(c)
	if err != nil {
		return nil
	}
	policies := enforcer.GetPolicy()
	ret := []*ormapi.RolePerm{}
	for ii, _ := range policies {
		if len(policies[ii]) < 3 {
			continue
		}
		perm := ormapi.RolePerm{
			Role:     policies[ii][0],
			Resource: policies[ii][1],
			Action:   policies[ii][2],
		}
		ret = append(ret, &perm)
	}
	return c.JSON(http.StatusOK, ret)
}

// Show roles assigned to the current user
func ShowRoleAssignment(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return nil
	}

	super := false
	if enforcer.Enforce(claims.Username, "", ResourceUsers, ActionView) {
		// super user, show all roles
		super = true
	}

	groupings := enforcer.GetGroupingPolicy()
	ret := []*ormapi.Role{}
	for ii, _ := range groupings {
		role := parseRole(groupings[ii])
		if role == nil {
			continue
		}
		if !super && claims.Username != role.Username {
			continue
		}
		ret = append(ret, role)
	}
	return c.JSON(http.StatusOK, ret)
}

func parseRole(grp []string) *ormapi.Role {
	if len(grp) < 2 {
		return nil
	}
	role := ormapi.Role{Role: grp[1]}
	domuser := strings.Split(grp[0], "::")
	if len(domuser) > 1 {
		role.Org = domuser[0]
		role.Username = domuser[1]
	} else {
		role.Username = grp[0]
	}
	return &role
}

func getCasbinGroup(org, username string) string {
	return org + "::" + username
}

func ShowRole(c echo.Context) error {
	rolemap := make(map[string]struct{})
	policies := enforcer.GetPolicy()
	for _, policy := range policies {
		if len(policy) < 1 {
			continue
		}
		rolemap[policy[0]] = struct{}{}
	}
	roles := make([]string, 0)
	for role, _ := range rolemap {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return c.JSON(http.StatusOK, roles)
}

func AddUserRole(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	role := ormapi.Role{}
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	err = AddUserRoleObj(claims, &role)
	return setReply(c, err, Msg("Role added to user"))
}

func AddUserRoleObj(claims *UserClaims, role *ormapi.Role) error {
	if role.Username == "" {
		return fmt.Errorf("Username not specified")
	}
	if role.Org == "" {
		return fmt.Errorf("Organziation not specified")
	}
	if role.Role == "" {
		return fmt.Errorf("Role not specified")
	}
	// check that user/org/role exists
	res := db.Where(&ormapi.User{Name: role.Username}).First(&ormapi.User{})
	if res.RecordNotFound() {
		return fmt.Errorf("Username not found")
	}
	if res.Error != nil {
		return dbErr(res.Error)
	}
	res = db.Where(&ormapi.Organization{Name: role.Org}).First(&ormapi.Organization{})
	if res.RecordNotFound() {
		return fmt.Errorf("Organization not found")
	}
	if res.Error != nil {
		return dbErr(res.Error)
	}
	policies := enforcer.GetPolicy()
	roleFound := false
	for _, policy := range policies {
		if len(policy) < 1 {
			continue
		}
		if policy[0] == role.Role {
			roleFound = true
			break
		}
	}
	if !roleFound {
		return fmt.Errorf("Role not found")
	}

	// make sure caller has perms to modify users of target org
	if !enforcer.Enforce(claims.Username, role.Org, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	psub := getCasbinGroup(role.Org, role.Username)
	enforcer.AddGroupingPolicy(psub, role.Role)
	return nil
}

func RemoveUserRole(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	role := ormapi.Role{}
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	err = RemoveUserRoleObj(claims, &role)
	return setReply(c, err, Msg("Role removed from user"))
}

func RemoveUserRoleObj(claims *UserClaims, role *ormapi.Role) error {
	if role.Username == "" {
		return fmt.Errorf("Username not specified")
	}
	if role.Org == "" {
		return fmt.Errorf("Organziation not specified")
	}
	if role.Role == "" {
		return fmt.Errorf("Role not specified")
	}

	// make sure caller has perms to modify users of target org
	if !enforcer.Enforce(claims.Username, role.Org, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}

	psub := getCasbinGroup(role.Org, role.Username)
	enforcer.RemoveGroupingPolicy(psub, role.Role)
	return nil
}

func ShowUserRole(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	roles, err := ShowUserRoleObj(claims)
	return setReply(c, err, roles)
}

// show roles for organizations the current user has permission to
// add/remove roles to. This "shows" all the actions taken by
// Add/RemoveUserRole.
func ShowUserRoleObj(claims *UserClaims) ([]ormapi.Role, error) {
	roles := []ormapi.Role{}

	groupings := enforcer.GetGroupingPolicy()
	for ii, _ := range groupings {
		role := parseRole(groupings[ii])
		if role == nil {
			continue
		}
		if !enforcer.Enforce(claims.Username, role.Org, ResourceUsers, ActionView) {
			continue
		}
		roles = append(roles, *role)
	}
	return roles, nil
}

// for debugging
func dumpRbac() {
	policies := enforcer.GetPolicy()
	for _, p := range policies {
		fmt.Printf("policy: %+v\n", p)
	}
	groups := enforcer.GetGroupingPolicy()
	for _, grp := range groups {
		fmt.Printf("group: %+v\n", grp)
	}
}
