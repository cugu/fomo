package db

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "modernc.org/sqlite" // import sqlite driver

	"github.com/cugu/fomo/db/sqlc"
)

const sqlitePath = "./fomo.db"

//go:embed migrations/*.sql
var schema embed.FS

func DB() (*sql.DB, *sqlc.Queries, error) {
	name, err := filepath.Abs(sqlitePath)
	if err != nil {
		return nil, nil, err
	}

	sqlite, err := sql.Open("sqlite", name+"?_time_format=sqlite")
	if err != nil {
		return nil, nil, err
	}

	// create table if not exists
	u, _ := url.Parse(fmt.Sprintf("sqlite://%s", name))
	db := dbmate.New(u)
	db.FS = schema
	db.MigrationsTableName = "migrations"
	db.MigrationsDir = []string{"migrations"}

	return sqlite, sqlc.New(sqlite), db.CreateAndMigrate()
}
