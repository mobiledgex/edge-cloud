package orm

import (
	"fmt"
	"strings"
	"time"

	gormadapter "github.com/casbin/gorm-adapter"
	"github.com/jinzhu/gorm"
	_ "github.com/labstack/echo"
	_ "github.com/lib/pq"
	"github.com/mobiledgex/edge-cloud/log"
)

var retryInterval = 10 * time.Second

func InitSql(addr, username, password, dbname string) (*gorm.DB, *gormadapter.Adapter, error) {
	hostport := strings.Split(addr, ":")
	if len(hostport) != 2 {
		return nil, nil, fmt.Errorf("Invalid postgres address format %s", addr)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"dbname=%s sslmode=disable password=%s",
		hostport[0], hostport[1], username, dbname, password)
	var err error
	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		log.InfoLog("init sql", "host", hostport[0], "port", hostport[1],
			"dbname", dbname)
		return nil, nil, err
	}

	dbSpecified := true
	adapter := gormadapter.NewAdapter("postgres", psqlInfo, dbSpecified)

	return db, adapter, nil
}

func InitData(superuser, superpass string, stop *bool, done chan struct{}) {
	if db == nil {
		log.FatalLog("db not initialized")
	}
	first := true
	for {
		if *stop {
			return
		}
		if !first {
			time.Sleep(retryInterval)
		}
		first = false

		// create or update tables
		err := db.AutoMigrate(&User{}, &Organization{}, &Controller{}).Error
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "automigrate", "err", err)
			continue
		}
		// create initial database data
		err = InitRolePerms()
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "init roles", "err", err)
			continue
		}
		err = InitAdmin(superuser, superpass)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "init admin", "err", err)
			continue
		}
		log.DebugLog(log.DebugLevelApi, "init data done")
		break
	}
	close(done)
}
