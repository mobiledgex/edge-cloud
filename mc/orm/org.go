package orm

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

// Organization Type names for ORM database
var OrgTypeAdmin = "admin"
var OrgTypeDeveloper = "developer"
var OrgTypeOperator = "operator"

func CreateOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	// any user can create their own organization
	org := Organization{}
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if org.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("Name not specified"))
	}
	if strings.Contains(org.Name, "::") {
		return c.JSON(http.StatusBadRequest, Msg("Name cannot contain ::"))
	}

	role := ""
	if org.Type == OrgTypeDeveloper {
		role = RoleDeveloperManager
	} else if org.Type == OrgTypeOperator {
		role = RoleOperatorManager
	} else {
		return c.JSON(http.StatusBadRequest, Msg(fmt.Sprintf("Organization type must be %s, or %s", OrgTypeDeveloper, OrgTypeOperator)))
	}
	if org.Address == "" {
		return c.JSON(http.StatusBadRequest, Msg("Address not specified"))
	}
	if org.Phone == "" {
		return c.JSON(http.StatusBadRequest, Msg("Phone number not specified"))
	}
	org.AdminUserID = claims.UserID
	err = db.Create(&org).Error
	if err != nil {
		return err
	}
	// set user to admin role of organization
	psub := getCasbinGroup(org.Name, claims.UserID)
	enforcer.AddGroupingPolicy(psub, role)
	return c.JSON(http.StatusOK, MsgName("Organization created", org.Name))
}

func DeleteOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	org := &Organization{}
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if org.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("Organization name not specified"))
	}
	if !enforcer.Enforce(id2str(claims.UserID), org.Name, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	// delete org
	err = db.Delete(&org).Error
	if err != nil {
		return err
	}
	// delete all casbin groups associated with org
	groups := enforcer.GetGroupingPolicy()
	for _, grp := range groups {
		if len(grp) < 2 {
			continue
		}
		strs := strings.Split(grp[0], "::")
		if len(strs) == 2 && strs[0] == org.Name {
			enforcer.RemoveGroupingPolicy(grp[0], grp[1])
		}
	}
	return c.JSON(http.StatusOK, Msg("Organization deleted"))
}

// Show Organizations that current user belongs to.
func ShowOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}

	orgs := []Organization{}
	if enforcer.Enforce(id2str(claims.UserID), "", ResourceUsers, ActionView) {
		// super user, show all orgs
		err = db.Find(&orgs).Error
	} else {
		// show orgs for current user
		userIDStr := strconv.FormatInt(claims.UserID, 10)
		groupings := enforcer.GetGroupingPolicy()
		for _, grp := range groupings {
			if len(grp) < 2 {
				continue
			}
			orguser := strings.Split(grp[0], "::")
			if len(orguser) > 1 && orguser[1] == userIDStr {
				org := Organization{}
				org.Name = orguser[0]
				err = db.Where(&org).First(&org).Error
				if err != nil {
					return err
				}
				orgs = append(orgs, org)
			}
		}
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, orgs)
}
