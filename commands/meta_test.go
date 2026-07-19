package commands

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("version")
	require.NoError(t, err)
	assert.Contains(t, out, "torre")
}

func TestVersion_JSON(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("version", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"version"`)
}

func TestConfigPathAndSetUse(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, "config.yaml")

	_, _, err = e.run("config", "set", "api_base_url", "https://torre.ai/api", "--profile", "work")
	require.NoError(t, err)

	_, _, err = e.run("config", "use", "work")
	require.NoError(t, err)

	out, _, err = e.run("config", "view")
	require.NoError(t, err)
	assert.Contains(t, out, "current_profile: work")

	out, _, err = e.run("config", "list-profiles")
	require.NoError(t, err)
	assert.Contains(t, out, "work")
}

func TestConfigSet_UnknownKey(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("config", "set", "nope", "x")
	require.Error(t, err)
}

func TestAuthLoginLogoutStatus(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("auth", "login", "--token", "abc123")
	require.NoError(t, err)
	tok, err := e.store.Get("default")
	require.NoError(t, err)
	assert.Equal(t, "abc123", tok)

	out, _, err := e.run("auth", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "stored")

	_, _, err = e.run("auth", "logout")
	require.NoError(t, err)
	_, gerr := e.store.Get("default")
	require.Error(t, gerr)

	out, _, err = e.run("auth", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "none")
}

func TestAliasSetListRemove(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("alias", "set", "go", "jobs search --skill golang")
	require.NoError(t, err)
	out, _, err := e.run("alias", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "go = jobs search --skill golang")
	_, _, err = e.run("alias", "remove", "go")
	require.NoError(t, err)
}

func TestAlias_CannotShadowBuiltin(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("alias", "set", "jobs", "genome x")
	require.Error(t, err)
}

func TestExpandAliases_BuiltinWins(t *testing.T) {
	// A built-in name is never expanded even if an alias exists.
	got := ExpandAliases([]string{"jobs", "search"})
	assert.Equal(t, []string{"jobs", "search"}, got)
}

func TestDoctor(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("doctor")
	require.NoError(t, err)
	assert.Contains(t, out, "connectivity")
	assert.Contains(t, out, "search OK")
}

func TestDoctor_JSON_FailOnError(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{}`))
	})
	_, _, err := e.run("doctor", "--json")
	require.Error(t, err)
}

func TestInit(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("init")
	require.NoError(t, err)
	assert.Contains(t, out, "setup")
}

func TestAPI_RawGet(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"person":{"name":"x"}}`))
	out, _, err := e.run("api", "GET", "genome/bios/x", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "name")
}

func TestAPI_InvalidMethod(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("api", "FETCH", "x")
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "invalid method")
}

func TestAPI_SearchHost(t *testing.T) {
	e := newEnv(t, jsonHandler(`{"total":0,"results":[]}`))
	out, _, err := e.run("api", "POST", "opportunities/_search/", "--host", "search", "-d", `{}`, "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "total")
}

func TestCompletion(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("completion", "bash")
	require.NoError(t, err)
	assert.Contains(t, out, "torre")
}

func TestUnknownOutputFormat(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	_, _, err := e.run("jobs", "search", "--skill", "go", "-o", "xml")
	require.Error(t, err)
}
