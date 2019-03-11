package orm

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/util"
)

// Init admin creates the admin user and adds the admin role.
func InitAdmin(superuser, superpass string) error {
	log.DebugLog(log.DebugLevelApi, "init admin")

	// create superuser if it doesn't exist
	passhash, salt, iter := NewPasshash(superpass)
	super := ormapi.User{
		Name:          superuser,
		Email:         superuser + "@mobiledgex.net",
		EmailVerified: true,
		Passhash:      passhash,
		Salt:          salt,
		Iter:          iter,
		GivenName:     superuser,
		FamilyName:    superuser,
		Nickname:      superuser,
	}
	err := db.FirstOrCreate(&super, &ormapi.User{Name: superuser}).Error
	if err != nil {
		return err
	}

	// set role of superuser to admin manager
	enforcer.AddGroupingPolicy(super.Name, RoleAdminManager)
	return nil
}

var BadAuthDelay = time.Second

func Login(c echo.Context) error {
	login := ormapi.UserLogin{}
	if err := c.Bind(&login); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if login.Username == "" {
		return c.JSON(http.StatusBadRequest, Msg("Username not specified"))
	}
	user := ormapi.User{}
	lookup := ormapi.User{Name: login.Username}
	err := db.Where(&lookup).First(&user).Error
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "user lookup failed", "lookup", lookup, "err", err)
		time.Sleep(BadAuthDelay)
		return c.JSON(http.StatusBadRequest, Msg("Invalid username or password"))
	}
	matches, err := PasswordMatches(login.Password, user.Passhash, user.Salt, user.Iter)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "password matches err", "err", err)
	}
	if !matches || err != nil {
		time.Sleep(BadAuthDelay)
		return c.JSON(http.StatusBadRequest, Msg("Invalid username or password"))
	}
	cookie, err := GenerateCookie(&user)
	if err != nil {
		log.DebugLog(log.DebugLevelApi, "failed to generate cookie", "err", err)
		return c.JSON(http.StatusBadRequest, Msg("Failed to generate cookie"))
	}
	return c.JSON(http.StatusOK, M{"token": cookie})
}

func CreateUser(c echo.Context) error {
	user := ormapi.User{}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if user.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("Name not specified"))
	}
	if strings.Contains(user.Name, "::") {
		return c.JSON(http.StatusBadRequest, Msg("Name cannot contain ::"))
	}
	if !util.ValidEmail(user.Email) {
		return c.JSON(http.StatusBadRequest, Msg("Invalid email address"))
	}
	if err := ValidPassword(user.Passhash); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid password, "+
			err.Error()))
	}
	if !util.ValidLDAPName(user.Name) {
		return c.JSON(http.StatusBadRequest, Msg("Invalid characters in user name"))
	}
	user.EmailVerified = false
	// password should be passed through in Passhash field.
	user.Passhash, user.Salt, user.Iter = NewPasshash(user.Passhash)
	if err := db.Create(&user).Error; err != nil {
		return setReply(c, dbErr(err), nil)
	}
	gitlabCreateLDAPUser(&user)

	return c.JSON(http.StatusOK, Msg("user created"))
}

func DeleteUser(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	user := ormapi.User{}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if user.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("User Name not specified"))
	}
	// Only user themself or super-user can delete user.
	if user.Name != claims.Username && !enforcer.Enforce(claims.Username, "", ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	// delete role mappings
	groups := enforcer.GetGroupingPolicy()
	for _, grp := range groups {
		if len(grp) < 2 {
			continue
		}
		strs := strings.Split(grp[0], "::")
		if grp[0] == user.Name || (len(strs) == 2 && strs[1] == user.Name) {
			enforcer.RemoveGroupingPolicy(grp[0], grp[1])
		}
	}
	// delete user
	err = db.Delete(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	gitlabDeleteLDAPUser(user.Name)

	return c.JSON(http.StatusOK, Msg("user deleted"))
}

// Show current user info
func CurrentUser(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	user := ormapi.User{Name: claims.Username}
	err = db.Where(&user).First(&user).Error
	if err != nil {
		return setReply(c, dbErr(err), nil)
	}
	user.Passhash = ""
	user.Salt = ""
	user.Iter = 0
	return c.JSON(http.StatusOK, user)
}

// Show users by Organization
func ShowUser(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	filter := ormapi.Organization{}
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	users := []ormapi.User{}
	if !enforcer.Enforce(claims.Username, filter.Name, ResourceUsers, ActionView) {
		return echo.ErrForbidden
	}
	// if filter ID is 0, show all users (super user only)
	if filter.Name == "" {
		err = db.Find(&users).Error
		if err != nil {
			return setReply(c, dbErr(err), nil)
		}
	} else {
		groupings := enforcer.GetGroupingPolicy()
		for _, grp := range groupings {
			if len(grp) < 2 {
				continue
			}
			orguser := strings.Split(grp[0], "::")
			if len(orguser) > 1 && orguser[0] == filter.Name {
				user := ormapi.User{}
				user.Name = orguser[1]
				err = db.Where(&user).First(&user).Error
				if err != nil {
					return setReply(c, dbErr(err), nil)
				}
				users = append(users, user)
			}
		}
	}
	for ii, _ := range users {
		// don't show auth/private info
		users[ii].Passhash = ""
		users[ii].Salt = ""
		users[ii].Iter = 0
	}
	return c.JSON(http.StatusOK, users)
}
