package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sinceSearchBody has three opportunities spanning years so a date filter is observable:
// a 2021 job (must drop), a 2025 job and a 2026 job (must keep for a 2022+ threshold).
const sinceSearchBody = `{"total":3,"size":20,"results":[` +
	`{"id":"old2021","objective":"Ancient Go role","created":"2021-03-01T10:00:00.000Z"},` +
	`{"id":"keep2025","objective":"Go role 2025","created":"2025-05-01T10:00:00.000Z"},` +
	`{"id":"recent","objective":"Fresh Go role","created":"2026-07-15T08:30:00.000Z"}]}`

func TestJobsSearch_Since_AbsoluteDropsOld(t *testing.T) {
	e := newEnv(t, jsonHandler(sinceSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--since", "2022-01-01", "-o", "id")
	require.NoError(t, err)
	assert.NotContains(t, out, "old2021")
	assert.Contains(t, out, "keep2025")
	assert.Contains(t, out, "recent")
}

func TestJobsSearch_PostedAfterAlias(t *testing.T) {
	e := newEnv(t, jsonHandler(sinceSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--posted-after", "2026-01-01", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "recent\n", out)
}

func TestJobsSearch_Since_WithExplicitLimit(t *testing.T) {
	// An explicit --limit must suppress the auto-widened scan and still filter by date.
	e := newEnv(t, jsonHandler(sinceSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--since", "2022-01-01", "--limit", "5", "-o", "id")
	require.NoError(t, err)
	assert.NotContains(t, out, "old2021")
	assert.Contains(t, out, "keep2025")
}

func TestJobsSearch_Since_Invalid(t *testing.T) {
	e := newEnv(t, jsonHandler(sinceSearchBody))
	_, _, err := e.run("jobs", "search", "--skill", "go", "--since", "yesterday")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --since value")
}

func TestJobsSearch_Since_Relative(t *testing.T) {
	// A far-back relative window keeps everything from the fixture (all newer than 9999 days ago).
	e := newEnv(t, jsonHandler(sinceSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--since", "9999d", "-o", "id")
	require.NoError(t, err)
	assert.Contains(t, out, "old2021")
	assert.Contains(t, out, "recent")
}

func TestJobsSearch_Since_DryRunValidates(t *testing.T) {
	e := newEnv(t, jsonHandler(sinceSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--since", "7d", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X POST")

	_, _, err = e.run("jobs", "search", "--skill", "go", "--since", "bad", "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --since value")
}
