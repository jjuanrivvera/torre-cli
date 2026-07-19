package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		value   string
		want    time.Time
		wantErr bool
	}{
		{"empty is zero", "", time.Time{}, false},
		{"absolute date", "2026-07-12", time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC), false},
		{"relative days", "7d", now.Add(-7 * 24 * time.Hour), false},
		{"relative single day", "1d", now.Add(-24 * time.Hour), false},
		{"relative weeks", "2w", now.Add(-14 * 24 * time.Hour), false},
		{"relative zero days", "0d", now, false},
		{"invalid word", "yesterday", time.Time{}, true},
		{"invalid unit", "7m", time.Time{}, true},
		{"invalid slashed date", "2026/07/12", time.Time{}, true},
		{"invalid negative", "-7d", time.Time{}, true},
		{"invalid trailing", "7dd", time.Time{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSince(tc.value, now)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseSince(%q) expected error, got nil", tc.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSince(%q) unexpected error: %v", tc.value, err)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("ParseSince(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestFilterByCreated(t *testing.T) {
	raw := func(id, created string) json.RawMessage {
		if created == "" {
			return json.RawMessage(`{"id":"` + id + `"}`)
		}
		return json.RawMessage(`{"id":"` + id + `","created":"` + created + `"}`)
	}
	results := []json.RawMessage{
		raw("old2021", "2021-03-01T10:00:00.000Z"),
		raw("recent", "2026-07-15T08:30:00.000Z"),
		raw("boundary", "2026-07-12T00:00:00.000Z"),
		raw("missing", ""),
		json.RawMessage(`{"id":"garbage","created":"not-a-date"}`),
	}
	threshold := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)

	got := FilterByCreated(results, threshold)

	ids := map[string]bool{}
	for _, r := range got {
		var env struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(r, &env); err != nil {
			t.Fatalf("unmarshal kept result: %v", err)
		}
		ids[env.ID] = true
	}
	if ids["old2021"] {
		t.Error("2021 opportunity should have been dropped")
	}
	if !ids["recent"] {
		t.Error("recent opportunity should have been kept")
	}
	if !ids["boundary"] {
		t.Error("opportunity created exactly at the threshold should be kept (inclusive)")
	}
	if ids["missing"] {
		t.Error("opportunity with no created timestamp should be dropped")
	}
	if ids["garbage"] {
		t.Error("opportunity with an unparseable created should be dropped")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 kept results, got %d", len(got))
	}
}

func TestFilterByCreated_ZeroThresholdNoOp(t *testing.T) {
	results := []json.RawMessage{
		json.RawMessage(`{"id":"a","created":"2021-01-01T00:00:00.000Z"}`),
	}
	got := FilterByCreated(results, time.Time{})
	if len(got) != 1 {
		t.Fatalf("zero threshold should return input unchanged, got %d", len(got))
	}
}
