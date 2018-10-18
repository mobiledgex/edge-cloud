// Unit test support for local postgres process
package main

import (
	"database/sql"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/integration/process"
)

var PostgresUnitTestAddr = "127.0.0.1:5432"

type PostgresServer struct {
	process process.SqlLocal
}

func NewPostgresServer() *PostgresServer {
	return &PostgresServer{
		process: process.SqlLocal{
			Name:     "unit-test-postgres",
			DataDir:  "/var/tmp/.postgres",
			HttpAddr: PostgresUnitTestAddr,
		},
	}
}

// Ensure Postgres SQL database for Unit Test is running and is clean
// (removes tables)
func EnsureCleanSql() (*sql.DB, error) {
	s := NewPostgresServer()
	// see if sql database is already running
	out, _ := exec.Command("pg_ctl", "-D", s.process.DataDir, "status").CombinedOutput()
	if !strings.Contains(string(out), "server is running") {
		err := s.Start()
		if err != nil {
			return nil, err
		}
	}
	db, err := InitSql(s.process.HttpAddr)
	if err != nil {
		return nil, err
	}
	err = sqlResetDB(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	err = InitSqlTables(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, err
}

// Same as EnsureCleanSql but when caller doesn't need db handle
func EnsureCleanSql2() error {
	db, err := EnsureCleanSql()
	if err == nil {
		db.Close()
	}
	return err
}

func (s *PostgresServer) Start() error {
	return s.process.Start("")
}

func (s *PostgresServer) Stop() {
	s.process.Stop()
}

func sqlResetDB(db *sql.DB) error {
	stmt := "DROP TABLE IF EXISTS Operators, Developers, Users, Roles CASCADE"
	_, err := db.Exec(stmt)
	return err
}
