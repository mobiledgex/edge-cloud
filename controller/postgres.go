package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

var InitTableRetryInterval = 10 * time.Second

func InitSql(addr string) (*sql.DB, error) {
	hostport := strings.Split(addr, ":")
	if len(hostport) != 2 {
		return nil, fmt.Errorf("Invalid postgres address format %s", addr)
	}

	password := ""
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"dbname=%s sslmode=disable password=%s",
		hostport[0], hostport[1], cloudcommon.PostgresUserName,
		cloudcommon.PostgresUserDb, password)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	if *shortTimeouts {
		InitTableRetryInterval = 100 * time.Millisecond
	}
	return db, nil
}

// This function is called by UnitTest code so we don't
// have to async wait for go thread in GoInitSqlTables().
func InitSqlTables(db *sql.DB) error {
	_, err := edgeproto.CreateOperatorSqlTable(db)
	if err != nil {
		return err
	}
	_, err = edgeproto.CreateDeveloperSqlTable(db)
	if err != nil {
		return err
	}
	_, err = edgeproto.CreateUserSqlTable(db)
	if err != nil {
		return err
	}
	_, err = edgeproto.CreateRoleSqlTable(db)
	if err != nil {
		return err
	}
	return nil
}

func GoInitSqlTables(db *sql.DB) {
	// This is done as a separate thread since we don't know
	// when Sql database will be reachable.
	go func() {
		// make sure database has tables
		first := true
		for {
			if !first {
				time.Sleep(InitTableRetryInterval)
			}
			err := InitSqlTables(db)
			if err == nil {
				break
			}
			first = false
		}
	}()
}
