package app

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/cugu/fomo/db/sqlc"
	"github.com/cugu/fomo/feed"
)

func scheduleUpdates(config *Config, queries *sqlc.Queries) (func() error, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("error creating scheduler: %w", err)
	}

	if len(config.ReleaseTimes) == 0 {
		slog.Info("Automatic feed fetching disabled because no release times are configured")
		return scheduler.Shutdown, nil
	}

	slog.Info(
		"Scheduling feed fetches",
		"interval", config.FetchInterval,
		"release_times", config.ReleaseTimes,
	)

	if _, err := scheduler.NewJob(
		gocron.DurationJob(config.FetchInterval),
		gocron.NewTask(func() {
			availableAt := nextReleaseTime(time.Now(), config.ReleaseTimes)
			for _, f := range config.Feeds {
				if err := feed.Fetch(context.Background(), queries, f, &availableAt); err != nil {
					slog.Error("error fetching feed", "name", f.Name(), "error", err.Error())
				}
			}
		}),
		gocron.WithStartAt(gocron.WithStartImmediately()),
	); err != nil {
		return nil, fmt.Errorf("error creating job: %w", err)
	}

	scheduler.Start()

	return scheduler.Shutdown, nil
}

func nextReleaseTime(now time.Time, releaseTimes []int) time.Time {
	times := slices.Clone(releaseTimes)
	slices.Sort(times)

	for _, hour := range times {
		candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
		if !candidate.Before(now) {
			return candidate
		}
	}

	return time.Date(now.Year(), now.Month(), now.Day()+1, times[0], 0, 0, 0, now.Location())
}
