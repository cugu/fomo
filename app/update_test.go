package app

import (
	"testing"
	"time"
)

func TestNextReleaseTime(t *testing.T) {
	t.Parallel()

	zone := time.FixedZone("test", 2*60*60)
	tests := []struct {
		name  string
		now   time.Time
		times []int
		want  time.Time
	}{
		{
			name:  "later slot today",
			now:   time.Date(2026, 7, 24, 8, 30, 0, 0, zone),
			times: []int{16, 7},
			want:  time.Date(2026, 7, 24, 16, 0, 0, 0, zone),
		},
		{
			name:  "exact slot",
			now:   time.Date(2026, 7, 24, 16, 0, 0, 0, zone),
			times: []int{7, 16},
			want:  time.Date(2026, 7, 24, 16, 0, 0, 0, zone),
		},
		{
			name:  "first slot tomorrow",
			now:   time.Date(2026, 7, 24, 20, 0, 0, 0, zone),
			times: []int{16, 7},
			want:  time.Date(2026, 7, 25, 7, 0, 0, 0, zone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := nextReleaseTime(tt.now, tt.times); !got.Equal(tt.want) {
				t.Errorf("nextReleaseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
