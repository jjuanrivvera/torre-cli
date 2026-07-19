package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// filterSearchBody carries a mix of locationTypes and disclosed/undisclosed compensation so
// the client-side --location-type / --remote-anywhere / --comp-disclosed-only filters are
// observable end to end.
const filterSearchBody = `{"total":5,"size":20,"results":[` +
	`{"id":"anywhere-paid","objective":"Go anywhere","place":{"locationType":"remote_anywhere"},"compensation":{"data":{"minAmount":1200,"minHourlyUSD":7.5}}},` +
	`{"id":"anywhere-unpaid","objective":"Go anywhere no pay","place":{"locationType":"remote_anywhere"},"compensation":{"data":null}},` +
	`{"id":"tz-paid","objective":"Go timezone","place":{"locationType":"remote_timezones"},"compensation":{"data":{"minAmount":0,"minHourlyUSD":9}}},` +
	`{"id":"onsite","objective":"Go onsite","place":{"locationType":"on_site"},"compensation":{"data":{"minAmount":5000,"minHourlyUSD":0}}},` +
	`{"id":"noplace","objective":"Go noplace"}]}`

func TestJobsSearch_RemoteAnywhere(t *testing.T) {
	e := newEnv(t, jsonHandler(filterSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--remote-anywhere", "-o", "id")
	require.NoError(t, err)
	assert.Contains(t, out, "anywhere-paid")
	assert.Contains(t, out, "anywhere-unpaid")
	assert.NotContains(t, out, "tz-paid")
	assert.NotContains(t, out, "onsite")
	assert.NotContains(t, out, "noplace")
}

func TestJobsSearch_LocationTypeCSV(t *testing.T) {
	e := newEnv(t, jsonHandler(filterSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--location-type", "remote_anywhere,remote_timezones", "-o", "id")
	require.NoError(t, err)
	assert.Contains(t, out, "anywhere-paid")
	assert.Contains(t, out, "tz-paid")
	assert.NotContains(t, out, "onsite")
	assert.NotContains(t, out, "noplace")
}

func TestJobsSearch_LocationTypeCaseInsensitive(t *testing.T) {
	e := newEnv(t, jsonHandler(filterSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--location-type", "REMOTE_ANYWHERE", "-o", "id")
	require.NoError(t, err)
	assert.Contains(t, out, "anywhere-paid")
	assert.NotContains(t, out, "tz-paid")
}

func TestJobsSearch_CompDisclosedOnly(t *testing.T) {
	e := newEnv(t, jsonHandler(filterSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--comp-disclosed-only", "-o", "id")
	require.NoError(t, err)
	assert.Contains(t, out, "anywhere-paid")
	assert.Contains(t, out, "tz-paid")
	assert.Contains(t, out, "onsite")
	assert.NotContains(t, out, "anywhere-unpaid")
	assert.NotContains(t, out, "noplace")
}

func TestJobsSearch_RemoteAnywhereAndCompDisclosed(t *testing.T) {
	// The two hard filters compose (AND): only remote_anywhere AND disclosed pay survives.
	e := newEnv(t, jsonHandler(filterSearchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--remote-anywhere", "--comp-disclosed-only", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "anywhere-paid\n", out)
}
