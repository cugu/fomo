package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-co-op/gocron/v2"

	"github.com/cugu/fomo/db/sqlc"
	"github.com/cugu/fomo/feed"
)

func scheduleUpdates(config *Config, queries *sqlc.Queries) (func() error, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("error creating scheduler: %w", err)
	}

	var atTimes []gocron.AtTime

	for _, updateTime := range config.UpdateTimes {
		if updateTime < 0 || updateTime >= 24 {
			return nil, fmt.Errorf("invalid update time: %d", updateTime)
		}

		atTimes = append(atTimes, gocron.NewAtTime(uint(updateTime), 0, 0))
	}

	if len(atTimes) == 0 {
		return scheduler.Shutdown, nil
	}

	if _, err := scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(atTimes[0], atTimes[1:]...)),
		gocron.NewTask(func() {
			for name, f := range config.Feeds {
				if err := feed.Fetch(context.Background(), queries, f); err != nil {
					slog.Error("error fetching feed", "name", name, "error", err.Error())
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
