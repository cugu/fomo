package app

import (
	"reflect"
	"testing"
	"time"
)

func TestJSONConfigToConfig(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name         string
		jsonConfig   JSONConfig
		wantTimes    []int
		wantInterval time.Duration
		wantErr      bool
	}{
		{
			name:         "defaults to hourly fetching",
			jsonConfig:   JSONConfig{ReleaseTimes: []int{7, 16}},
			wantTimes:    []int{7, 16},
			wantInterval: time.Hour,
		},
		{
			name:         "accepts deprecated update times",
			jsonConfig:   JSONConfig{UpdateTimes: []int{8, 20}},
			wantTimes:    []int{8, 20},
			wantInterval: time.Hour,
		},
		{
			name: "release times take precedence",
			jsonConfig: JSONConfig{
				ReleaseTimes:         []int{9},
				UpdateTimes:          []int{10},
				FetchIntervalMinutes: 15,
			},
			wantTimes:    []int{9},
			wantInterval: 15 * time.Minute,
		},
		{
			name:       "rejects invalid release time",
			jsonConfig: JSONConfig{ReleaseTimes: []int{24}},
			wantErr:    true,
		},
		{
			name:       "rejects negative fetch interval",
			jsonConfig: JSONConfig{FetchIntervalMinutes: -1},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := tt.jsonConfig.toConfig()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}

				return
			}

			if err != nil {
				t.Fatalf("toConfig() error = %v", err)
			}

			if !reflect.DeepEqual(config.ReleaseTimes, tt.wantTimes) {
				t.Errorf("ReleaseTimes = %v, want %v", config.ReleaseTimes, tt.wantTimes)
			}

			if config.FetchInterval != tt.wantInterval {
				t.Errorf("FetchInterval = %v, want %v", config.FetchInterval, tt.wantInterval)
			}
		})
	}
}
