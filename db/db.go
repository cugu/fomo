package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "modernc.org/sqlite" // import sqlite driver

	"github.com/cugu/fomo/db/sqlc"
)

const sqlitePath = "fomo.db"

//go:embed migrations/*.sql
var schema embed.FS

func DB(dataDirPath string) (*sql.DB, *sqlc.Queries, error) {
	dbPath := filepath.Join(dataDirPath, sqlitePath)

	slog.Info("Connecting to database", "path", dbPath)

	sqlite, err := sql.Open("sqlite", dbPath+"?_time_format=sqlite")
	if err != nil {
		return nil, nil, err
	}

	// create table if not exists
	u, _ := url.Parse(fmt.Sprintf("sqlite://%s", dbPath))
	db := dbmate.New(u)
	db.FS = schema
	db.MigrationsTableName = "migrations"
	db.MigrationsDir = []string{"migrations"}
	db.Log = &infoSlog{}

	return sqlite, sqlc.New(sqlite), db.CreateAndMigrate()
}

type infoSlog struct{}

func (i *infoSlog) Write(p []byte) (n int, err error) {
	slog.Info(string(p))

	return len(p), nil
}
