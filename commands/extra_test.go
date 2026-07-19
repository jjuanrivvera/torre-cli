package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/torre-cli/internal/update"
	"github.com/jjuanrivvera/torre-cli/internal/version"
)

func TestUpdateCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9","assets":[]}`))
	}))
	defer srv.Close()
	old := newUpdater
	newUpdater = func() *update.Updater { return update.NewUpdaterWithBaseURL(version.Version, srv.URL) }
	t.Cleanup(func() { newUpdater = old })

	e := newEnv(t, nil)
	out, _, err := e.run("update", "check")
	require.NoError(t, err)
	assert.Contains(t, out, "Latest:  v9.9.9")
}

func TestUpdate_DevBuildNoop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9","assets":[]}`))
	}))
	defer srv.Close()
	old := newUpdater
	newUpdater = func() *update.Updater { return update.NewUpdaterWithBaseURL("dev", srv.URL) }
	t.Cleanup(func() { newUpdater = old })

	e := newEnv(t, nil)
	out, _, err := e.run("update")
	require.NoError(t, err)
	assert.Contains(t, out, "Already on the latest version.")
}

func TestVersion_Check(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.0.1"}`))
	}))
	defer srv.Close()
	old := latestReleaseURL
	latestReleaseURL = srv.URL
	t.Cleanup(func() { latestReleaseURL = old })

	e := newEnv(t, nil)
	out, _, err := e.run("version", "--check")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestReadDataArg_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "body.json")
	require.NoError(t, os.WriteFile(f, []byte(`{"x":1}`), 0o600))
	e := newEnv(t, jsonHandler(`{"ok":true}`))
	// api command reads --data @file
	out, _, err := e.run("api", "POST", "opportunities/_search/", "--host", "search", "-d", "@"+f, "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "ok")
}

func TestAliasExpansion_RealTree(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	_, _, err := e.run("alias", "set", "g", "jobs search --skill golang")
	require.NoError(t, err)
	// Now expand through the real config.
	got := ExpandAliases([]string{"g", "--remote"})
	assert.Equal(t, []string{"jobs", "search", "--skill", "golang", "--remote"}, got)
	// Unknown token is passed through.
	assert.Equal(t, []string{"nope"}, ExpandAliases([]string{"nope"}))
	assert.Equal(t, []string(nil), ExpandAliases(nil))
}

func TestPromptSecret_PipedStdin(t *testing.T) {
	d := newDeps()
	root := newRootCmd(d)
	sub, _, err := root.Find([]string{"version"})
	require.NoError(t, err)
	sub.SetIn(strings.NewReader("piped-token\n"))
	got, err := promptSecret(sub, "token: ")
	require.NoError(t, err)
	assert.Equal(t, "piped-token", got)
}
