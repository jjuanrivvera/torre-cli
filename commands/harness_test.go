package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/jjuanrivvera/torre-cli/internal/api"
	"github.com/jjuanrivvera/torre-cli/internal/auth"
)

// fakeStore is an in-memory auth.Store so tests never touch a real OS keyring.
type fakeStore struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeStore() *fakeStore { return &fakeStore{data: map[string]string{}} }

func (f *fakeStore) Set(profile, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[profile] = token
	return nil
}

func (f *fakeStore) Get(profile string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if t, ok := f.data[profile]; ok && t != "" {
		return t, nil
	}
	return "", auth.ErrNotFound
}

func (f *fakeStore) Delete(profile string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, profile)
	return nil
}

func (f *fakeStore) Backend() string { return "fake" }

// env wires one test invocation: an httptest Torre server, an isolated config dir, and a
// fake keyring.
type env struct {
	t     *testing.T
	srv   *httptest.Server
	store *fakeStore
}

// newEnv starts a mock Torre server (both hosts route to it) and isolates state under
// t.TempDir().
func newEnv(t *testing.T, handler http.HandlerFunc) *env {
	t.Helper()
	e := &env{t: t, store: newFakeStore()}
	if handler != nil {
		e.srv = httptest.NewServer(handler)
		t.Cleanup(e.srv.Close)
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("TORRE_PROFILE", "")
	t.Setenv("TORRE_TOKEN", "")
	t.Setenv("TORRE_BASE_URL", "")
	t.Setenv("TORRE_SEARCH_BASE_URL", "")
	t.Setenv("NO_COLOR", "1")
	return e
}

func (e *env) deps() *deps {
	d := newDeps()
	d.store = func() auth.Store { return e.store }
	if e.srv != nil {
		url := e.srv.URL
		httpc := e.srv.Client()
		// Ignore the resolved bases and force both hosts at the mock server.
		d.newClient = func(_, _ string, opts ...api.Option) *api.Client {
			opts = append(opts, api.WithHTTPClient(httpc), api.WithMaxRetries(0))
			return api.NewClientWithBaseURL(url, opts...)
		}
	}
	return d
}

// run executes the real command tree with captured output.
func (e *env) run(args ...string) (string, string, error) {
	e.t.Helper()
	return runWithDeps(e.t, e.deps(), args...)
}

// runWithDeps builds a fresh tree from d and runs it with captured output.
func runWithDeps(t *testing.T, d *deps, args ...string) (string, string, error) {
	t.Helper()
	var out, errB bytes.Buffer
	d.out = &out // dry-run curls go here too, so they are captured
	root := newRootCmd(d)
	root.SetArgs(args)
	root.SetOut(&out)
	root.SetErr(&errB)
	err := root.ExecuteContext(t.Context())
	return out.String(), errB.String(), err
}

// jsonHandler answers every request with one body.
func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

// routeHandler routes by URL path prefix so a single server serves search + app-API paths.
func routeHandler(routes map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		for prefix, body := range routes {
			if len(r.URL.Path) >= len(prefix) && r.URL.Path[:len(prefix)] == prefix {
				_, _ = w.Write([]byte(body))
				return
			}
		}
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error":"Not Found"}`))
	}
}
