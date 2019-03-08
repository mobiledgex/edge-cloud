package orm

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
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
	org := ormapi.Organization{}
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	err = CreateOrgObj(claims, &org)
	return setReply(c, err, MsgName("Organization created", org.Name))
}

func CreateOrgObj(claims *UserClaims, org *ormapi.Organization) error {
	if org.Name == "" {
		return fmt.Errorf("Name not specified")
	}
	if strings.Contains(org.Name, "::") {
		return fmt.Errorf("Name cannot contain ::")
	}
	if !ValidUsername(org.Name) {
		return fmt.Errorf("Invalid characters in name")
	}
	// any user can create their own organization

	role := ""
	if org.Type == OrgTypeDeveloper {
		role = RoleDeveloperManager
	} else if org.Type == OrgTypeOperator {
		role = RoleOperatorManager
	} else {
		return fmt.Errorf(fmt.Sprintf("Organization type must be %s, or %s", OrgTypeDeveloper, OrgTypeOperator))
	}
	if org.Address == "" {
		return fmt.Errorf("Address not specified")
	}
	if org.Phone == "" {
		return fmt.Errorf("Phone number not specified")
	}
	org.AdminUsername = claims.Username
	err := db.Create(&org).Error
	if err != nil {
		return dbErr(err)
	}
	// set user to admin role of organization
	psub := getCasbinGroup(org.Name, claims.Username)
	enforcer.AddGroupingPolicy(psub, role)

	gitlabCreateGroup(org)
	r := ormapi.Role{
		Org:      org.Name,
		Username: claims.Username,
		Role:     role,
	}
	gitlabAddGroupMember(&r)
	return nil
}

func DeleteOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	org := ormapi.Organization{}
	if err := c.Bind(&org); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	err = DeleteOrgObj(claims, &org)
	return setReply(c, err, Msg("Organization deleted"))
}

func DeleteOrgObj(claims *UserClaims, org *ormapi.Organization) error {
	if org.Name == "" {
		return fmt.Errorf("Organization name not specified")
	}
	if !enforcer.Enforce(claims.Username, org.Name, ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	// delete org
	err := db.Delete(&org).Error
	if err != nil {
		return dbErr(err)
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
	gitlabDeleteGroup(org)
	return nil
}

// Show Organizations that current user belongs to.
func ShowOrg(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	orgs, err := ShowOrgObj(claims)
	return setReply(c, err, orgs)
}

func ShowOrgObj(claims *UserClaims) ([]ormapi.Organization, error) {
	orgs := []ormapi.Organization{}
	if enforcer.Enforce(claims.Username, "", ResourceUsers, ActionView) {
		// super user, show all orgs
		err := db.Find(&orgs).Error
		if err != nil {
			return nil, dbErr(err)
		}
	} else {
		// show orgs for current user
		groupings := enforcer.GetGroupingPolicy()
		for _, grp := range groupings {
			if len(grp) < 2 {
				continue
			}
			orguser := strings.Split(grp[0], "::")
			if len(orguser) > 1 && orguser[1] == claims.Username {
				org := ormapi.Organization{}
				org.Name = orguser[0]
				err := db.Where(&org).First(&org).Error
				if err != nil {
					return nil, dbErr(err)
				}
				orgs = append(orgs, org)
			}
		}
	}
	return orgs, nil
}
