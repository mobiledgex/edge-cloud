package orm

import (
	"fmt"
	"net/http"
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
	defer func() {
		if err != nil {
			db.Delete(&org)
		}
	}()
	// set user to admin role of organization
	psub := getCasbinGroup(org.ID, claims.UserID)
	enforcer.AddGroupingPolicy(psub, role)
	defer func() {
		if err != nil {
			enforcer.RemoveGroupingPolicy(psub, role)
		}
	}()
	// add user as org member
	userOrg := UserOrg{
		UserID: claims.UserID,
		OrgID:  org.ID,
	}
	err = db.Create(&userOrg).Error
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, MsgID("Organization created", org.ID))
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
	if org.ID == 0 {
		return c.JSON(http.StatusBadRequest, Msg("Organization ID not specified"))
	}
	if !enforcer.Enforce(id2str(claims.UserID), org.ID, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	userOrg := UserOrg{OrgID: org.ID}
	// delete member associations
	err = db.Where(&userOrg).Delete(UserOrg{}).Error
	if err != nil {
		return err
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
		if len(strs) == 2 && strs[0] == id2str(org.ID) {
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
		err = db.Joins("JOIN user_orgs on user_orgs.org_id=organizations.id").
			Where("user_orgs.user_id=?", claims.UserID).Find(&orgs).Error
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, orgs)
}
