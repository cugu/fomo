package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cugu/fomo/feed"
)

type Config struct {
	BaseURL       string
	Port          int
	Feeds         []feed.Feed
	ReleaseTimes  []int
	FetchInterval time.Duration
}

const defaultFetchInterval = time.Hour

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
	BaseURL              string                     `json:"base_url"`
	Password             string                     `json:"password"`
	Port                 int                        `json:"port"`
	Feeds                map[string]json.RawMessage `json:"feeds"`
	ReleaseTimes         []int                      `json:"release_times"`
	UpdateTimes          []int                      `json:"update_times"`
	FetchIntervalMinutes int                        `json:"fetch_interval_minutes"`
}

type TypedConfig struct {
	Type string `json:"type"`
}

func (j *JSONConfig) toConfig() (*Config, error) {
	feeds, err := j.parseFeeds()
	if err != nil {
		return nil, err
	}

	releaseTimes, err := j.parseReleaseTimes()
	if err != nil {
		return nil, err
	}

	fetchInterval, err := j.parseFetchInterval()
	if err != nil {
		return nil, err
	}

	return &Config{
		BaseURL:       j.BaseURL,
		Port:          j.Port,
		Feeds:         feeds,
		ReleaseTimes:  releaseTimes,
		FetchInterval: fetchInterval,
	}, nil
}

func (j *JSONConfig) parseFeeds() ([]feed.Feed, error) {
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

	return feeds, nil
}

func (j *JSONConfig) parseReleaseTimes() ([]int, error) {
	releaseTimes := j.ReleaseTimes
	if releaseTimes == nil {
		releaseTimes = j.UpdateTimes

		if j.UpdateTimes != nil {
			slog.Warn("update_times is deprecated; use release_times instead")
		}
	} else if j.UpdateTimes != nil {
		slog.Warn("ignoring deprecated update_times because release_times is configured")
	}

	for _, releaseTime := range releaseTimes {
		if releaseTime < 0 || releaseTime >= 24 {
			return nil, fmt.Errorf("invalid release time: %d", releaseTime)
		}
	}

	return releaseTimes, nil
}

func (j *JSONConfig) parseFetchInterval() (time.Duration, error) {
	if j.FetchIntervalMinutes < 0 {
		return 0, fmt.Errorf("invalid fetch interval: %d minutes", j.FetchIntervalMinutes)
	}

	if j.FetchIntervalMinutes > 0 {
		return time.Duration(j.FetchIntervalMinutes) * time.Minute, nil
	}

	return defaultFetchInterval, nil
}
