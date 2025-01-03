package app

import (
	"fmt"

	"github.com/cugu/fomo/db"
	"github.com/cugu/fomo/server"
)

// Run starts the server.
func Run() error {
	config, err := parseConfig()
	if err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	database, queries, err := db.DB()
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	defer database.Close()

	teardownScheduler, err := scheduleUpdates(config, queries)
	if err != nil {
		return fmt.Errorf("error scheduling updates: %w", err)
	}
	defer teardownScheduler()

	fomoServer := server.New(config.BaseURL, config.Password, config.UpdateTimes, config.Feeds, queries)

	return fomoServer.ListenAndServe(config.Port)
}
