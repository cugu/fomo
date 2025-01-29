package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/cugu/fomo/feed"
)

type Config struct {
	BaseURL     string
	Port        int
	Feeds       []feed.Feed
	UpdateTimes []int
}

func parseConfig(configPath string) (*Config, error) {
	slog.Info("Loading config", "path", configPath)

	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	var cfg JSONConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg.toConfig()
}

type JSONConfig struct {
	BaseURL     string                     `json:"base_url"`
	Password    string                     `json:"password"`
	Port        int                        `json:"port"`
	Feeds       map[string]json.RawMessage `json:"feeds"`
	UpdateTimes []int                      `json:"update_times"`
}

type TypedConfig struct {
	Type string `json:"type"`
}

func (j *JSONConfig) toConfig() (*Config, error) {
	var feeds []feed.Feed

	for name, raw := range j.Feeds {
		var typed TypedConfig
		if err := json.Unmarshal(raw, &typed); err != nil {
			return nil, err
		}

		generator, ok := feed.LookupFeed(typed.Type)
		if !ok {
			return nil, fmt.Errorf("unknown feed type %s", typed.Type)
		}

		feed, err := generator(name, raw)
		if err != nil {
			return nil, fmt.Errorf("error creating feed %s: %w", name, err)
		}

		feeds = append(feeds, feed)
	}

	return &Config{
		BaseURL:     j.BaseURL,
		Port:        j.Port,
		Feeds:       feeds,
		UpdateTimes: j.UpdateTimes,
	}, nil
}
