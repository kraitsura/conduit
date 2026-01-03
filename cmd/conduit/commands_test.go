package main

import (
	"testing"
)

func TestHighlightMatch(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		query  string
		expect string
	}{
		{
			name:   "exact match at start",
			text:   "hello world",
			query:  "hello",
			expect: colorYellow + "hello" + colorReset + colorDim + " world",
		},
		{
			name:   "match in middle",
			text:   "the quick brown",
			query:  "quick",
			expect: "the " + colorYellow + "quick" + colorReset + colorDim + " brown",
		},
		{
			name:   "case insensitive match",
			text:   "Hello World",
			query:  "hello",
			expect: colorYellow + "Hello" + colorReset + colorDim + " World",
		},
		{
			name:   "no match",
			text:   "hello world",
			query:  "foo",
			expect: "hello world",
		},
		{
			name:   "match at end",
			text:   "hello world",
			query:  "world",
			expect: "hello " + colorYellow + "world" + colorReset + colorDim + "",
		},
		{
			name:   "empty query",
			text:   "hello",
			query:  "",
			expect: colorYellow + "" + colorReset + colorDim + "hello",
		},
		{
			name:   "empty text",
			text:   "",
			query:  "foo",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highlightMatch(tt.text, tt.query)
			if got != tt.expect {
				t.Errorf("highlightMatch(%q, %q) = %q, want %q", tt.text, tt.query, got, tt.expect)
			}
		})
	}
}
