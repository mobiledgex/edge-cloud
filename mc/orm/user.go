package orm

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/util"
)

// Init admin creates the admin user and adds the admin role.
func InitAdmin(superuser, superpass string) error {
	log.DebugLog(log.DebugLevelApi, "init admin")

	// create superuser if it doesn't exist
	passhash, salt, iter := NewPasshash(superpass)
	super := User{
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
	err := db.FirstOrCreate(&super, &User{Name: superuser}).Error
	if err != nil {
		return err
	}

	// set role of superuser to admin manager
	enforcer.AddGroupingPolicy(id2str(super.ID), RoleAdminManager)
	return nil
}

type UserLogin struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

var BadAuthDelay = time.Second

func Login(c echo.Context) error {
	login := UserLogin{}
	if err := c.Bind(&login); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if login.Username == "" {
		return c.JSON(http.StatusBadRequest, Msg("Username not specified"))
	}
	user := User{}
	lookup := User{Name: login.Username}
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
		return c.JSON(http.StatusBadRequest, Msg("Failed to generate cookie"))
	}
	return c.JSON(http.StatusOK, M{"token": cookie})
}

func CreateUser(c echo.Context) error {
	user := User{}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if user.Name == "" {
		return c.JSON(http.StatusBadRequest, Msg("Name not specified"))
	}
	if !util.ValidEmail(user.Email) {
		return c.JSON(http.StatusBadRequest, Msg("Invalid email address"))
	}
	if err := ValidPassword(user.Passhash); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid password, "+
			err.Error()))
	}
	user.EmailVerified = false
	user.ID = 0 // auto-assigned
	// password should be passed through in Passhash field.
	user.Passhash, user.Salt, user.Iter = NewPasshash(user.Passhash)
	if err := db.Create(&user).Error; err != nil {
		log.DebugLog(log.DebugLevelApi, "create user", "user", user,
			"err", err)
		return c.JSON(http.StatusBadRequest, Msg("User create failed"))
	}
	return c.JSON(http.StatusOK, MsgID("user created", user.ID))
}

func DeleteUser(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	user := User{}
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	if user.ID == 0 {
		return c.JSON(http.StatusBadRequest, Msg("User ID not specified"))
	}
	// Only user themself or super-user can delete user.
	if user.ID != claims.UserID && !enforcer.Enforce(id2str(claims.UserID), "", ResourceUsers, ActionManage) {
		return echo.ErrForbidden
	}
	// delete org associations
	userOrg := UserOrg{UserID: user.ID}
	err = db.Where(&userOrg).Delete(UserOrg{}).Error
	if err != nil {
		return err
	}
	// delete role mappings
	groups := enforcer.GetGroupingPolicy()
	for _, grp := range groups {
		if len(grp) < 2 {
			continue
		}
		strs := strings.Split(grp[0], "::")
		if grp[0] == id2str(user.ID) || (len(strs) == 2 && strs[1] == id2str(user.ID)) {
			enforcer.RemoveGroupingPolicy(grp[0], grp[1])
		}
	}
	// delete user
	err = db.Delete(&user).Error
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Msg("user deleted"))
}

// Show current user info
func CurrentUser(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	user := User{ID: claims.UserID}
	err = db.Where(&user).First(&user).Error
	if err != nil {
		return err
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
	filter := Organization{}
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	users := []User{}
	if !enforcer.Enforce(id2str(claims.UserID), filter.ID, ResourceUsers, ActionView) {
		return echo.ErrForbidden
	}
	// if filter ID is 0, show all users (super user only)
	if filter.ID == 0 {
		err = db.Find(&users).Error
	} else {
		err = db.Joins("JOIN user_orgs on user_orgs.user_id=users.id").
			Where("user_orgs.org_id=?", filter.ID).Find(&users).Error
	}
	if err != nil {
		return err
	}
	for ii, _ := range users {
		// don't show auth/private info
		users[ii].Passhash = ""
		users[ii].Salt = ""
		users[ii].Iter = 0
	}
	return c.JSON(http.StatusOK, users)
}
