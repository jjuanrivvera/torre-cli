package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// sinceRelative matches the relative --since shorthand: an integer followed by a `d` (days)
// or `w` (weeks) unit, e.g. "7d" or "2w".
var sinceRelative = regexp.MustCompile(`^([0-9]+)([dw])$`)

// ParseSince interprets a --since / --posted-after value as an inclusive lower bound on an
// opportunity's `.created` timestamp. It accepts EITHER an absolute date "YYYY-MM-DD" OR a
// relative shorthand "Nd"/"Nw" (N days/weeks before now). An empty value returns the zero
// time (meaning "no lower bound"). `now` is injected so the relative form is deterministic
// under test.
func ParseSince(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if m := sinceRelative.FindStringSubmatch(value); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil { // the regex guarantees digits, but keep the contract explicit
			return time.Time{}, sinceError(value)
		}
		days := n
		if m[2] == "w" {
			days = n * 7
		}
		return now.Add(-time.Duration(days) * 24 * time.Hour), nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, sinceError(value)
}

func sinceError(value string) error {
	return fmt.Errorf(
		"invalid --since value %q: use an absolute date YYYY-MM-DD (e.g. 2026-07-12) "+
			"or a relative shorthand Nd/Nw (e.g. 7d, 2w)", value)
}

// createdEnvelope pulls just the created timestamp out of an opportunity result.
type createdEnvelope struct {
	Created string `json:"created"`
}

// FilterByCreated keeps only the results whose `.created` timestamp is at or after
// threshold. A zero threshold is a no-op (the input is returned unchanged). Results with a
// missing or unparseable `.created` are dropped: the point of a date filter is "posted
// on/after", and an item that can't be date-verified fails that test. Torre's search orders
// by relevance rather than date, so this client-side pass is the only date filter available
// (the search API silently ignores a created clause — see DECISIONS.md).
func FilterByCreated(results []json.RawMessage, threshold time.Time) []json.RawMessage {
	if threshold.IsZero() {
		return results
	}
	kept := make([]json.RawMessage, 0, len(results))
	for _, r := range results {
		var env createdEnvelope
		if err := json.Unmarshal(r, &env); err != nil || env.Created == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, env.Created)
		if err != nil {
			continue
		}
		if !t.Before(threshold) {
			kept = append(kept, r)
		}
	}
	return kept
}

// placeEnvelope pulls just the place.locationType out of an opportunity result. Torre models
// a remote role's openness here: `remote_anywhere` (open to any country) vs
// `remote_timezones` (restricted to timezone bands), etc.
type placeEnvelope struct {
	Place struct {
		LocationType string `json:"locationType"`
	} `json:"place"`
}

// FilterByLocationType keeps only the results whose `.place.locationType` matches one of the
// given values (case-insensitive). An empty `values` slice is a no-op. Because the point of
// the filter is "this role is open to my situation", a result with a missing or empty
// locationType is DROPPED when the filter is active (it can't be verified to match). This is
// an honest hard client-side filter — unlike the server-side `--location` ranking hint
// (DECISIONS.md #16, #18).
func FilterByLocationType(results []json.RawMessage, values []string) []json.RawMessage {
	if len(values) == 0 {
		return results
	}
	want := make(map[string]struct{}, len(values))
	for _, v := range values {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			want[v] = struct{}{}
		}
	}
	if len(want) == 0 {
		return results
	}
	kept := make([]json.RawMessage, 0, len(results))
	for _, r := range results {
		var env placeEnvelope
		if err := json.Unmarshal(r, &env); err != nil {
			continue
		}
		lt := strings.ToLower(strings.TrimSpace(env.Place.LocationType))
		if lt == "" {
			continue
		}
		if _, ok := want[lt]; ok {
			kept = append(kept, r)
		}
	}
	return kept
}

// compEnvelope pulls just the disclosed compensation figures out of an opportunity result.
// Most Torre opportunities leave compensation undisclosed (`.compensation.data` null, or the
// min figures zero); a disclosed one carries a positive minAmount or minHourlyUSD.
type compEnvelope struct {
	Compensation struct {
		Data *struct {
			MinAmount    float64 `json:"minAmount"`
			MinHourlyUSD float64 `json:"minHourlyUSD"`
		} `json:"data"`
	} `json:"compensation"`
}

// FilterByCompensationDisclosed keeps only the results that actually disclose a compensation
// figure (a positive minAmount OR minHourlyUSD). This is distinct from the server-side
// `--compensation` ranking hint (DECISIONS.md #16): that nudges ordering, this drops rows
// whose pay is undisclosed/zero. An item with no compensation data is dropped.
func FilterByCompensationDisclosed(results []json.RawMessage) []json.RawMessage {
	kept := make([]json.RawMessage, 0, len(results))
	for _, r := range results {
		var env compEnvelope
		if err := json.Unmarshal(r, &env); err != nil {
			continue
		}
		d := env.Compensation.Data
		if d != nil && (d.MinAmount > 0 || d.MinHourlyUSD > 0) {
			kept = append(kept, r)
		}
	}
	return kept
}
