package orm

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/casbin/casbin"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/log"
)

// Server struct is just to track sql/db so we can stop them later.
type Server struct {
	config       ServerConfig
	sql          *process.SqlLocal
	db           *gorm.DB
	echo         *echo.Echo
	stopInitData bool
	initDataDone chan struct{}
}

type ServerConfig struct {
	ServAddr      string
	SqlAddr       string
	VaultAddr     string
	RunLocal      bool
	InitLocal     bool
	IgnoreEnv     bool
	RbacModelPath string
}

var DefaultDBUser = "mcuser"
var DefaultDBName = "mcdb"
var DefaultDBPass = ""
var DefaultSuperuser = "mexadmin"
var DefaultSuperpass = "mexadmin123"

var db *gorm.DB
var enforcer *casbin.Enforcer

func RunServer(config *ServerConfig) (*Server, error) {
	dbuser := os.Getenv("db_username")
	dbpass := os.Getenv("db_password")
	dbname := os.Getenv("db_name")
	superuser := os.Getenv("superuser")
	superpass := os.Getenv("superpass")
	if dbuser == "" || config.IgnoreEnv {
		dbuser = DefaultDBUser
	}
	if dbname == "" || config.IgnoreEnv {
		dbname = DefaultDBName
	}
	if dbpass == "" || config.IgnoreEnv {
		dbpass = DefaultDBPass
	}
	if superuser == "" || config.IgnoreEnv {
		superuser = DefaultSuperuser
	}
	if superpass == "" || config.IgnoreEnv {
		superpass = DefaultSuperpass
	}
	if config.RbacModelPath == "" {
		config.RbacModelPath = "./rbac.conf"
		err := createRbacModel(config.RbacModelPath)
		if err != nil {
			return nil, fmt.Errorf("create default rbac model failed, %s", err.Error())
		}
	}

	InitVault(config.VaultAddr)

	server := Server{config: *config}
	if config.RunLocal {
		sql := process.SqlLocal{
			Name:     "sql1",
			DataDir:  "./.postgres",
			HttpAddr: config.SqlAddr,
			Username: dbuser,
			Dbname:   dbname,
		}
		_, err := os.Stat(sql.DataDir)
		if config.InitLocal || os.IsNotExist(err) {
			sql.InitDataDir()
		}
		err = sql.Start("")
		if err != nil {
			return nil, fmt.Errorf("local sql start failed, %s",
				err.Error())
		}
		server.sql = &sql
	}

	initdb, adapter, err := InitSql(config.SqlAddr, dbuser, dbpass, dbname)
	if err != nil {
		return nil, fmt.Errorf("sql init failed, %s", err.Error())
	}
	db = initdb
	server.db = db

	server.initDataDone = make(chan struct{}, 1)
	go InitData(superuser, superpass, &server.stopInitData, server.initDataDone)

	e := echo.New()
	e.HideBanner = true
	server.echo = e

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.Use(logger)

	// login route
	root := "api/v1"
	e.POST(root+"/login", Login)
	// accessible routes
	e.POST(root+"/usercreate", CreateUser)
	// authenticated routes - jwt middleware
	auth := e.Group(root + "/auth")
	auth.Use(AuthCookie)
	// authenticated routes - rbac via casbin (false arg disables logging)
	enforcer = casbin.NewEnforcer(config.RbacModelPath, adapter, false)
	// authenticated routes - gorm router
	auth.POST("/user/show", ShowUser)
	auth.POST("/user/current", CurrentUser)
	auth.POST("/user/delete", DeleteUser)
	auth.POST("/role/assignment/show", ShowRoleAssignment)
	auth.POST("/role/perms/show", ShowRolePerms)
	auth.POST("/role/show", ShowRole)
	auth.POST("/role/adduser", AddUserRole)
	auth.POST("/role/removeuser", RemoveUserRole)
	auth.POST("/org/create", CreateOrg)
	auth.POST("/org/show", ShowOrg)
	auth.POST("/org/delete", DeleteOrg)
	auth.POST("/controller/create", CreateController)
	auth.POST("/controller/delete", DeleteController)
	auth.POST("/controller/show", ShowController)
	addControllerApis(auth)
	go func() {
		err := e.Start(config.ServAddr)
		if err != nil && err != http.ErrServerClosed {
			log.FatalLog("Failed to serve", "err", err)
		}
	}()
	return &server, nil
}

func (s *Server) WaitUntilReady() error {
	// wait until init data is done
	<-s.initDataDone
	// wait until server is online
	for ii := 0; ii < 10; ii++ {
		resp, err := http.Get("http://" + s.config.ServAddr)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for server ready")
}

func (s *Server) Stop() {
	s.stopInitData = true
	s.echo.Close()
	s.db.Close()
	if s.sql != nil {
		s.sql.Stop()
	}
}
