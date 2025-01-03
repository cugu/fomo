package app

import (
	"cmp"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cugu/fomo/db"
	"github.com/cugu/fomo/server"
)

// Run starts the server.
func Run() error {
	password, configPath, dataDirPath, err := parseFlags()
	if err != nil {
		return fmt.Errorf("error parsing flags: %w", err)
	}

	config, err := parseConfig(configPath)
	if err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	database, queries, err := db.DB(dataDirPath)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	defer database.Close()

	teardownScheduler, err := scheduleUpdates(config, queries)
	if err != nil {
		return fmt.Errorf("error scheduling updates: %w", err)
	}
	defer teardownScheduler()

	fomoServer := server.New(config.BaseURL, password, config.UpdateTimes, config.Feeds, queries)

	slog.Info("Starting server", "url", config.BaseURL)

	return fomoServer.ListenAndServe(config.Port)
}

func parseFlags() (password, configFile, dataDir string, err error) {
	flag.StringVar(
		&password,
		"password",
		os.Getenv("FOMO_PASSWORD"),
		"password for the server",
	)
	flag.StringVar(
		&configFile,
		"config",
		cmp.Or(os.Getenv("FOMO_CONFIG"), "config.json"),
		"path to the config file",
	)
	flag.StringVar(
		&dataDir,
		"data",
		cmp.Or(os.Getenv("FOMO_DATA_DIR"), "."),
		"path to the data directory",
	)

	flag.Parse()

	if password == "" {
		return "", "", "", errors.New("password is required, either set FOMO_PASSWORD or use the -password flag")
	}

	configPath, err := filepath.Abs(configFile)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get absolute path for config file: %w", err)
	}

	dataDirPath, err := filepath.Abs(dataDir)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get absolute path for database file: %w", err)
	}

	return password, configPath, dataDirPath, nil
}
