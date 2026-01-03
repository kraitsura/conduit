package main

import (
	"testing"
	"time"
)

func TestPadRight(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		length int
		want   string
	}{
		{
			name:   "shorter string",
			s:      "hello",
			length: 10,
			want:   "     hello",
		},
		{
			name:   "exact length",
			s:      "hello",
			length: 5,
			want:   "hello",
		},
		{
			name:   "longer string",
			s:      "hello world",
			length: 5,
			want:   "hello world",
		},
		{
			name:   "empty string",
			s:      "",
			length: 5,
			want:   "     ",
		},
		{
			name:   "zero length",
			s:      "hello",
			length: 0,
			want:   "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padRight(tt.s, tt.length)
			if got != tt.want {
				t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.length, got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a smaller", 1, 5, 1},
		{"b smaller", 5, 1, 1},
		{"equal", 3, 3, 3},
		{"negative", -5, 5, -5},
		{"both negative", -5, -1, -5},
		{"zero", 0, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := min(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "just now",
			time: now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "minutes ago",
			time: now.Add(-5 * time.Minute),
			want: "5m ago",
		},
		{
			name: "hours ago",
			time: now.Add(-3 * time.Hour),
			want: "3h ago",
		},
		{
			name: "days ago",
			time: now.Add(-2 * 24 * time.Hour),
			want: "2d ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTimeAgo(tt.time)
			if got != tt.want {
				t.Errorf("formatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}
