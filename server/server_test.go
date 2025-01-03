package server

import "testing"

func Test_formatTimes(t *testing.T) {
	t.Parallel()

	type args struct {
		updateTimes []int
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{updateTimes: []int{}},
			want: "",
		},
		{
			name: "single",
			args: args{updateTimes: []int{1}},
			want: "1:00AM",
		},
		{
			name: "double",
			args: args{updateTimes: []int{1, 13}},
			want: "1:00AM and 1:00PM",
		},
		{
			name: "multiple",
			args: args{updateTimes: []int{1, 2, 13}},
			want: "1:00AM, 2:00AM, and 1:00PM",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := formatTimes(tt.args.updateTimes); got != tt.want {
				t.Errorf("formatTimes() = %v, want %v", got, tt.want)
			}
		})
	}
}
