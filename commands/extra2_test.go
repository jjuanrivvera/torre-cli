package commands

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runInput drives the tree with a scripted stdin (for wizard prompts).
func (e *env) runInput(stdin string, args ...string) (string, string, error) {
	d := e.deps()
	var out, errB bytes.Buffer
	d.out = &out
	root := newRootCmd(d)
	root.SetArgs(args)
	root.SetOut(&out)
	root.SetErr(&errB)
	root.SetIn(strings.NewReader(stdin))
	err := root.ExecuteContext(e.t.Context())
	return out.String(), errB.String(), err
}

func TestInit_StoresTokenWhenAccepted(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.runInput("y\nmy-token\n", "init")
	require.NoError(t, err)
	assert.Contains(t, out, "Stored a token")
	tok, gerr := e.store.Get("default")
	require.NoError(t, gerr)
	assert.Equal(t, "my-token", tok)
}

func TestAuthLogin_PromptedToken(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.runInput("prompted-secret\n", "auth", "login")
	require.NoError(t, err)
	tok, gerr := e.store.Get("default")
	require.NoError(t, gerr)
	assert.Equal(t, "prompted-secret", tok)
}

func TestDoctor_JSON_Success(t *testing.T) {
	e := newEnv(t, jsonHandler(searchBody))
	out, _, err := e.run("doctor", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"name": "connectivity"`)
	assert.Contains(t, out, `"ok": true`)
}

func TestJobsSearch_AllPaginates(t *testing.T) {
	e := newEnv(t, routeHandler(map[string]string{
		"/opportunities/_search/": `{"total":1,"size":20,"results":[{"id":"only"}]}`,
	}))
	out, _, err := e.run("jobs", "search", "--skill", "go", "--all", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "only\n", out)
}

func TestAPI_EmptyBody(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	out, _, err := e.run("api", "GET", "genome/bios/x")
	require.NoError(t, err)
	assert.Contains(t, out, "HTTP 200")
}
