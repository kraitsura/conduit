package main

import (
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

func TestCalculateSessionsPerDay(t *testing.T) {
	now := time.Now().Truncate(24 * time.Hour)

	tests := []struct {
		name     string
		sessions []types.ConduitSession
		days     int
		wantLen  int
		wantSum  float64
	}{
		{
			name:     "empty sessions",
			sessions: []types.ConduitSession{},
			days:     7,
			wantLen:  7,
			wantSum:  0,
		},
		{
			name: "single session today",
			sessions: []types.ConduitSession{
				{StartTime: now.Add(time.Hour)},
			},
			days:    7,
			wantLen: 7,
			wantSum: 1,
		},
		{
			name: "sessions across multiple days",
			sessions: []types.ConduitSession{
				{StartTime: now.Add(time.Hour)},            // today
				{StartTime: now.Add(-24 * time.Hour)},      // yesterday
				{StartTime: now.Add(-24 * time.Hour)},      // yesterday (2nd)
				{StartTime: now.Add(-48 * time.Hour)},      // 2 days ago
			},
			days:    7,
			wantLen: 7,
			wantSum: 4,
		},
		{
			name: "session older than range excluded",
			sessions: []types.ConduitSession{
				{StartTime: now.Add(-10 * 24 * time.Hour)}, // 10 days ago
			},
			days:    7,
			wantLen: 7,
			wantSum: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSessionsPerDay(tt.sessions, tt.days)

			if len(result) != tt.wantLen {
				t.Errorf("calculateSessionsPerDay() len = %d, want %d", len(result), tt.wantLen)
			}

			sum := 0.0
			for _, v := range result {
				sum += v
			}
			if sum != tt.wantSum {
				t.Errorf("calculateSessionsPerDay() sum = %f, want %f", sum, tt.wantSum)
			}
		})
	}
}

func TestCalculateAgentDistribution(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		sessions []types.ConduitSession
		wantNil  bool
		wantKeys []string
	}{
		{
			name:     "empty sessions",
			sessions: []types.ConduitSession{},
			wantNil:  true,
		},
		{
			name: "sessions without chats",
			sessions: []types.ConduitSession{
				{StartTime: now, Chats: []types.AgentChat{}},
			},
			wantNil: true,
		},
		{
			name: "single agent type",
			sessions: []types.ConduitSession{
				{
					StartTime: now,
					Chats: []types.AgentChat{
						{AgentType: "claude", StartTime: now, EndTime: timePtr(now.Add(time.Hour))},
					},
				},
			},
			wantNil:  false,
			wantKeys: []string{"claude"},
		},
		{
			name: "multiple agent types",
			sessions: []types.ConduitSession{
				{
					StartTime: now,
					Chats: []types.AgentChat{
						{AgentType: "claude", StartTime: now, EndTime: timePtr(now.Add(time.Hour))},
						{AgentType: "cursor", StartTime: now, EndTime: timePtr(now.Add(30 * time.Minute))},
					},
				},
			},
			wantNil:  false,
			wantKeys: []string{"claude", "cursor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAgentDistribution(tt.sessions)

			if tt.wantNil {
				if result != nil {
					t.Errorf("calculateAgentDistribution() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Error("calculateAgentDistribution() = nil, want non-nil")
				return
			}

			for _, key := range tt.wantKeys {
				if _, exists := result[key]; !exists {
					t.Errorf("calculateAgentDistribution() missing key %q", key)
				}
			}

			// Verify percentages sum to 100
			total := 0.0
			for _, v := range result {
				total += v
			}
			if total < 99.9 || total > 100.1 {
				t.Errorf("calculateAgentDistribution() percentages sum = %f, want ~100", total)
			}
		})
	}
}

func TestAgentColor(t *testing.T) {
	tests := []struct {
		agent string
		want  string
	}{
		{"claude", colorCyan},
		{"CLAUDE", colorCyan},
		{"Claude", colorCyan},
		{"cursor", colorGreen},
		{"aider", colorYellow},
		{"copilot", colorBlue},
		{"continue", "\033[35m"},
		{"cody", "\033[95m"},
		{"unknown", colorDim},
		{"", colorDim},
	}

	for _, tt := range tests {
		t.Run(tt.agent, func(t *testing.T) {
			got := agentColor(tt.agent)
			if got != tt.want {
				t.Errorf("agentColor(%q) = %q, want %q", tt.agent, got, tt.want)
			}
		})
	}
}

// timePtr returns a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}
