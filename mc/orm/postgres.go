package orm

import (
	"fmt"
	"strings"
	"time"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	gormadapter "github.com/casbin/gorm-adapter"
	"github.com/jinzhu/gorm"
	_ "github.com/labstack/echo"
	_ "github.com/lib/pq"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"go.uber.org/zap"
)

var retryInterval = 10 * time.Second

func InitSql(addr, username, password, dbname string) (*gorm.DB, persist.Adapter, error) {
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
			"dbname", dbname, "err", err)
		return nil, nil, err
	}

	dbSpecified := true
	adapter := gormadapter.NewAdapter("postgres", psqlInfo, dbSpecified)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	db.SetLogger(&sqlLogger{logger.Sugar()})
	db.LogMode(true)

	return db, &adapterLogger{adapter}, nil
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
		err := db.AutoMigrate(&ormapi.User{}, &ormapi.Organization{},
			&ormapi.Controller{}).Error
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

type sqlLogger struct {
	slog *zap.SugaredLogger
}

func (s *sqlLogger) Print(v ...interface{}) {
	if len(v) < 1 {
		return
	}
	kvs := make([]interface{}, 0)
	msg := "sql log"
	switch v[0] {
	case "sql":
		kvs = append(kvs, "sql")
		kvs = append(kvs, v[3])
		kvs = append(kvs, "vars")
		kvs = append(kvs, v[4])
		kvs = append(kvs, "rows-affected")
		kvs = append(kvs, v[5])
		kvs = append(kvs, "lineno")
		kvs = append(kvs, v[1])
		kvs = append(kvs, "took")
		kvs = append(kvs, v[2])
		msg = "Call sql"
	default:
		kvs = append(kvs, "vals")
		kvs = append(kvs, v[2:])
		kvs = append(kvs, "lineno")
		kvs = append(kvs, v[1])
	}
	log.DebugSLog(s.slog, log.DebugLevelApi, msg, kvs...)
}

type adapterLogger struct {
	adapter persist.Adapter
}

func (s *adapterLogger) LoadPolicy(model model.Model) error {
	start := time.Now()
	err := s.adapter.LoadPolicy(model)
	log.DebugLog(log.DebugLevelApi, "Call gorm LoadPolicy", "model", model, "took", time.Since(start))
	return err
}

func (s *adapterLogger) SavePolicy(model model.Model) error {
	start := time.Now()
	err := s.adapter.SavePolicy(model)
	log.DebugLog(log.DebugLevelApi, "Call gorm SavePolicy", "model", model, "took", time.Since(start))
	return err
}

func (s *adapterLogger) AddPolicy(sec, ptype string, rule []string) error {
	start := time.Now()
	err := s.adapter.AddPolicy(sec, ptype, rule)
	log.DebugLog(log.DebugLevelApi, "Call gorm AddPolicy", "sec", sec, "ptype", ptype, "rule", rule, "took", time.Since(start))
	return err
}

func (s *adapterLogger) RemovePolicy(sec, ptype string, rule []string) error {
	start := time.Now()
	err := s.adapter.RemovePolicy(sec, ptype, rule)
	log.DebugLog(log.DebugLevelApi, "Call gorm RemovePolicy", "sec", sec, "ptype", ptype, "rule", rule, "took", time.Since(start))
	return err
}

func (s *adapterLogger) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	start := time.Now()
	err := s.adapter.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
	log.DebugLog(log.DebugLevelApi, "Call gorm RemoveFilteredPolicy", "sec", sec, "ptype", ptype, "fieldIndex", fieldIndex, "fieldValues", fieldValues, "took", time.Since(start))
	return err
}
