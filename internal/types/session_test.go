package types

import (
	"testing"
	"time"
)

func TestConduitSessionDuration(t *testing.T) {
	tests := []struct {
		name     string
		session  ConduitSession
		wantMin  time.Duration
		wantMax  time.Duration
	}{
		{
			name: "completed session",
			session: func() ConduitSession {
				start := time.Now().Add(-2 * time.Hour)
				end := time.Now().Add(-1 * time.Hour)
				return ConduitSession{
					StartTime: start,
					EndTime:   &end,
				}
			}(),
			wantMin: 59 * time.Minute,
			wantMax: 61 * time.Minute,
		},
		{
			name: "active session",
			session: ConduitSession{
				StartTime: time.Now().Add(-30 * time.Minute),
				EndTime:   nil,
				IsActive:  true,
			},
			wantMin: 29 * time.Minute,
			wantMax: 31 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.Duration()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("Duration() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestConduitSessionTotalChatTime(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	thirtyMinAgo := now.Add(-30 * time.Minute)
	fifteenMinAgo := now.Add(-15 * time.Minute)

	session := ConduitSession{
		Chats: []AgentChat{
			{
				StartTime: oneHourAgo,
				EndTime:   &thirtyMinAgo, // 30 min
			},
			{
				StartTime: fifteenMinAgo,
				EndTime:   &now, // 15 min
			},
		},
	}

	got := session.TotalChatTime()
	want := 45 * time.Minute

	// Allow 1 second tolerance
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("TotalChatTime() = %v, want ~%v", got, want)
	}
}

func TestAgentChatDuration(t *testing.T) {
	tests := []struct {
		name    string
		chat    AgentChat
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			name: "completed chat",
			chat: func() AgentChat {
				start := time.Now().Add(-45 * time.Minute)
				end := time.Now().Add(-15 * time.Minute)
				return AgentChat{
					StartTime: start,
					EndTime:   &end,
				}
			}(),
			wantMin: 29 * time.Minute,
			wantMax: 31 * time.Minute,
		},
		{
			name: "active chat",
			chat: AgentChat{
				StartTime: time.Now().Add(-10 * time.Minute),
				IsActive:  true,
			},
			wantMin: 9 * time.Minute,
			wantMax: 11 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.chat.Duration()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("Duration() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSessionGap(t *testing.T) {
	if SessionGap != 30*time.Minute {
		t.Errorf("SessionGap = %v, want 30 minutes", SessionGap)
	}
}
