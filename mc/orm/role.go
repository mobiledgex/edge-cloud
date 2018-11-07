package orm

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/log"
)

const ActionView = "view"
const ActionManage = "manage"

const ResourceUsers = "users"
const ResourceApps = "apps"
const ResourceAppInsts = "appinsts"
const ResourceClusters = "clusters"
const ResourceClusterInsts = "clusterinsts"
const ResourceAppAnalytics = "appanalytics"
const ResourceCloudlets = "cloudlets"
const ResourceCloudletAnalytics = "cloudletanalytics"

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
	OrgID  int64
	UserID int64
	Role   string
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
		role.OrgID, _ = strconv.ParseInt(domuser[0], 10, 64)
		role.UserID, _ = strconv.ParseInt(domuser[1], 10, 64)
	} else {
		role.UserID, _ = strconv.ParseInt(grp[0], 10, 64)
	}
	return &role
}

func getCasbinGroup(orgID, userID int64) string {
	return id2str(orgID) + "::" + id2str(userID)
}

func id2str(id int64) string {
	return strconv.FormatInt(id, 10)
}
