package orm

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/log"
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

type RolePerm struct {
	Role     string
	Resource string
	Action   string
}

func ShowRolePerms(c echo.Context) error {
	_, err := getClaims(c)
	if err != nil {
		return nil
	}
	policies := enforcer.GetPolicy()
	ret := []*RolePerm{}
	for ii, _ := range policies {
		if len(policies[ii]) < 3 {
			continue
		}
		perm := RolePerm{
			Role:     policies[ii][0],
			Resource: policies[ii][1],
			Action:   policies[ii][2],
		}
		ret = append(ret, &perm)
	}
	return c.JSON(http.StatusOK, ret)
}

type Role struct {
	Org    string `form:"org" json:"org"`
	UserID int64  `form:"userid" json:"userid"`
	Role   string `form:"role" json:"role"`
}

// Show roles assigned to the current user
func ShowRoleAssignment(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return nil
	}

	super := false
	if enforcer.Enforce(id2str(claims.UserID), "", ResourceUsers, ActionView) {
		// super user, show all roles
		super = true
	}

	groupings := enforcer.GetGroupingPolicy()
	ret := []*Role{}
	for ii, _ := range groupings {
		role := parseRole(groupings[ii])
		if role == nil {
			continue
		}
		if !super && claims.UserID != role.UserID {
			continue
		}
		ret = append(ret, role)
	}
	return c.JSON(http.StatusOK, ret)
}

func parseRole(grp []string) *Role {
	if len(grp) < 2 {
		return nil
	}
	role := Role{Role: grp[1]}
	domuser := strings.Split(grp[0], "::")
	if len(domuser) > 1 {
		role.Org = domuser[0]
		role.UserID, _ = strconv.ParseInt(domuser[1], 10, 64)
	} else {
		role.UserID, _ = strconv.ParseInt(grp[0], 10, 64)
	}
	return &role
}

func getCasbinGroup(org string, userID int64) string {
	return org + "::" + id2str(userID)
}

func id2str(id int64) string {
	return strconv.FormatInt(id, 10)
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
	role := Role{}
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if role.UserID == 0 {
		return c.JSON(http.StatusBadRequest, Msg("UserID not specified"))
	}
	if role.Org == "" {
		return c.JSON(http.StatusBadRequest, Msg("Organziation not specified"))
	}
	if role.Role == "" {
		return c.JSON(http.StatusBadRequest, Msg("Role not specified"))
	}
	// check that user/org/role exists
	res := db.Where(&User{ID: role.UserID}).First(&User{})
	if res.RecordNotFound() {
		return c.JSON(http.StatusBadRequest, Msg("UserID not found"))
	}
	if res.Error != nil {
		return res.Error
	}
	res = db.Where(&Organization{Name: role.Org}).First(&Organization{})
	if res.RecordNotFound() {
		return c.JSON(http.StatusBadRequest, Msg("Organization not found"))
	}
	if res.Error != nil {
		return res.Error
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
		return c.JSON(http.StatusBadRequest, Msg("Role not found"))
	}

	// make sure caller has perms to modify users of target org
	if !enforcer.Enforce(id2str(claims.UserID), role.Org, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	psub := getCasbinGroup(role.Org, role.UserID)
	enforcer.AddGroupingPolicy(psub, role.Role)
	return c.JSON(http.StatusOK, Msg("Role added to user"))
}

func RemoveUserRole(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	role := Role{}
	if err := c.Bind(&role); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if role.UserID == 0 {
		return c.JSON(http.StatusBadRequest, Msg("UserID not specified"))
	}
	if role.Org == "" {
		return c.JSON(http.StatusBadRequest, Msg("Organziation not specified"))
	}
	if role.Role == "" {
		return c.JSON(http.StatusBadRequest, Msg("Role not specified"))
	}

	// make sure caller has perms to modify users of target org
	if !enforcer.Enforce(id2str(claims.UserID), role.Org, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}

	psub := getCasbinGroup(role.Org, role.UserID)
	enforcer.RemoveGroupingPolicy(psub, role.Role)
	return c.JSON(http.StatusOK, Msg("Role removed from user"))
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
