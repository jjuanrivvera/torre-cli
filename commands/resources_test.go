package commands

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const searchBody = `{"total":2,"size":20,"results":[{"id":"a","objective":"Go Engineer","remote":true},{"id":"b","objective":"Go Dev","remote":false}]}`

func TestJobsSearch_JSON(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Go Engineer")
	assert.Contains(t, out, `"id": "a"`)
}

func TestJobsSearch_ID(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "a\nb\n", out)
}

func TestJobsSearch_Table(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--columns", "id,objective")
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "Go Engineer")
}

func TestJobsSearch_SendsFilters(t *testing.T) {
	var gotBody string
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchBody))
	})
	_, _, err := e.run("jobs", "search", "--skill", "go", "--remote", "--location", "Colombia")
	require.NoError(t, err)
	assert.Contains(t, gotBody, `"remote":{"term":true}`)
	assert.Contains(t, gotBody, `"location":{"term":"Colombia"}`)
}

func TestJobsSearch_DryRun(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "_search")
}

func TestJobsGet(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"id":"KWN4QjAd","objective":"Go Engineer"}`))
	out, _, err := e.run("jobs", "get", "KWN4QjAd", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "KWN4QjAd")
}

func TestJobsGet_NotFound(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error":"Not Found"}`))
	})
	_, _, err := e.run("jobs", "get", "bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGenome(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"person":{"name":"Alex Torrenegra"}}`))
	out, _, err := e.run("genome", "torrenegra", "--jq", ".person.name")
	require.NoError(t, err)
	assert.Contains(t, out, "Alex Torrenegra")
}

func TestGenome_Yaml(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"person":{"name":"Alex"}}`))
	out, _, err := e.run("genome", "torrenegra", "-o", "yaml")
	require.NoError(t, err)
	assert.Contains(t, out, "name: Alex")
}

func TestPeopleSearch(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"total":1,"results":[{"name":"x","username":"u"}]}`))
	out, _, err := e.run("people", "search", "--skill", "go", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"username": "u"`)
}

func TestJobsSearch_ServerError(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"meta":{"message":"boom"}}`))
	})
	_, _, err := e.run("jobs", "search", "--skill", "go")
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "server error")
}

func TestUsesStoredToken(t *testing.T) {
	var gotAuth string
	e := newEnv(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchBody))
	})
	require.NoError(t, e.store.Set("default", "tok-abc"))
	_, _, err := e.run("jobs", "search", "--skill", "go")
	require.NoError(t, err)
	assert.Equal(t, "Bearer tok-abc", gotAuth)
}
