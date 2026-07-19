// Package api is the Torre.ai client core. Torre's public surface spans two hosts — the
// search cluster (search.torre.co, POST _search endpoints) and the app API
// (torre.ai/api, GET opportunity detail + genome). The client holds both bases and routes
// each typed method to the right one, over one shared request path with idempotent-only
// retry (honoring Retry-After), a dry-run curl mode, and an optional bearer token.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Default hosts for Torre's public API. Both are overridable (config/flags, and tests point
// them at a single httptest server that routes by path).
const (
	// DefaultSearchBaseURL serves the opportunities/people _search endpoints.
	DefaultSearchBaseURL = "https://search.torre.co"
	// DefaultAPIBaseURL serves opportunity detail (/suite/opportunities/{id}) and genome
	// (/genome/bios/{username}).
	DefaultAPIBaseURL = "https://torre.ai/api"
	// DefaultUserAgent is sent on every request. Torre's edge is friendlier to a
	// browser-like UA than to a bare Go client, so a sensible default keeps the public
	// (no-auth) path working out of the box.
	DefaultUserAgent = "torre-cli (+https://github.com/jjuanrivvera/torre-cli)"
)

// TokenFunc supplies an optional bearer token per request. Torre's public endpoints need
// none; a token is only used when the user configured one (some authenticated endpoints).
// It may be nil.
type TokenFunc func(ctx context.Context) (string, error)

// Client is a Torre.ai HTTP client spanning the search and app-API hosts.
type Client struct {
	searchBase string
	apiBase    string
	token      TokenFunc
	httpc      *http.Client
	userAgent  string

	// DryRun prints the equivalent curl to DryRunOut instead of sending the request.
	DryRun    bool
	DryRunOut io.Writer
	// ShowToken reveals the bearer token in dry-run output (redacted by default).
	ShowToken bool

	Verbose    bool
	VerboseOut io.Writer

	maxRetries int
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the HTTP transport (tests point it at httptest servers).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpc = h } }

// WithDryRun enables curl-printing mode.
func WithDryRun(dry bool, out io.Writer) Option {
	return func(c *Client) { c.DryRun = dry; c.DryRunOut = out }
}

// WithToken sets the optional bearer token source.
func WithToken(t TokenFunc) Option { return func(c *Client) { c.token = t } }

// WithUserAgent overrides the default User-Agent.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// WithMaxRetries overrides the retry budget (tests set 0 for speed).
func WithMaxRetries(n int) Option { return func(c *Client) { c.maxRetries = n } }

// New builds a Torre client for the two hosts. Empty bases fall back to the defaults.
func New(searchBase, apiBase string, opts ...Option) *Client {
	if searchBase == "" {
		searchBase = DefaultSearchBaseURL
	}
	if apiBase == "" {
		apiBase = DefaultAPIBaseURL
	}
	c := &Client{
		searchBase: strings.TrimRight(searchBase, "/"),
		apiBase:    strings.TrimRight(apiBase, "/"),
		httpc:      http.DefaultClient,
		userAgent:  DefaultUserAgent,
		DryRunOut:  os.Stdout,
		VerboseOut: os.Stderr,
		maxRetries: 3,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// NewClientWithBaseURL builds a client with both hosts pointed at the same base URL. Tests
// use it to drive every endpoint against one httptest server (routing by path), and it also
// backs a single-host --base-url override.
func NewClientWithBaseURL(base string, opts ...Option) *Client {
	return New(base, base, opts...)
}

// SearchBaseURL returns the resolved search host.
func (c *Client) SearchBaseURL() string { return c.searchBase }

// APIBaseURL returns the resolved app-API host.
func (c *Client) APIBaseURL() string { return c.apiBase }

// getJSON GETs a full URL and decodes the JSON body into out (nil discards it).
func (c *Client) getJSON(ctx context.Context, fullURL string, out any) (json.RawMessage, error) {
	status, _, body, err := c.doURL(ctx, http.MethodGet, fullURL, nil, nil)
	if err != nil {
		return nil, err
	}
	if status == 0 { // dry-run
		return nil, nil
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return nil, fmt.Errorf("decode Torre response: %w", err)
		}
	}
	return body, nil
}

// postJSON POSTs body to a full URL and returns the raw JSON response.
func (c *Client) postJSON(ctx context.Context, fullURL string, body []byte) (json.RawMessage, error) {
	status, _, resp, err := c.doURL(ctx, http.MethodPost, fullURL, body, nil)
	if err != nil {
		return nil, err
	}
	if status == 0 { // dry-run
		return nil, nil
	}
	return resp, nil
}

// Do sends one request against a chosen host base and returns status, headers, and body.
// host is "search" or "api"; path is relative to that host's base. A dry-run returns status
// 0. Non-2xx returns an *APIError. This is the raw escape hatch used by `torre api`.
func (c *Client) Do(ctx context.Context, host, method, path string, q url.Values, body []byte) (int, http.Header, []byte, error) {
	base := c.apiBase
	if host == "search" {
		base = c.searchBase
	}
	u := base + "/" + strings.TrimLeft(path, "/")
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return c.doURL(ctx, method, u, body, nil)
}

// doURL is the shared request path. It applies the User-Agent, optional bearer, retry, and
// dry-run.
func (c *Client) doURL(ctx context.Context, method, fullURL string, body []byte, extra map[string]string) (int, http.Header, []byte, error) {
	headers := map[string]string{
		"Accept":     "application/json",
		"User-Agent": c.userAgent,
	}
	if body != nil {
		headers["Content-Type"] = "application/json"
	}
	for k, v := range extra {
		if v == "" {
			delete(headers, k)
			continue
		}
		headers[k] = v
	}

	if c.DryRun {
		c.printCurl(ctx, method, fullURL, body, headers)
		return 0, nil, nil, nil
	}

	tok := ""
	if c.token != nil {
		t, err := c.token(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		tok = t
	}

	send := func() (*http.Response, error) {
		var rdr io.Reader
		if body != nil {
			rdr = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, rdr)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		if c.Verbose {
			fmt.Fprintf(c.VerboseOut, "> %s %s\n", method, fullURL)
		}
		return c.httpc.Do(req)
	}

	resp, err := c.sendWithRetry(ctx, method, send)
	if err != nil {
		return 0, nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return 0, nil, nil, fmt.Errorf("read response: %w", err)
	}
	if c.Verbose {
		fmt.Fprintf(c.VerboseOut, "< HTTP %d (%d bytes)\n", resp.StatusCode, len(respBody))
	}
	if resp.StatusCode >= 400 {
		return resp.StatusCode, resp.Header, respBody, parseAPIError(resp.StatusCode, respBody, resp.Header)
	}
	return resp.StatusCode, resp.Header, respBody, nil
}

// printCurl emits a copy-pasteable curl equivalent, redacting the token unless --show-token.
func (c *Client) printCurl(ctx context.Context, method, fullURL string, body []byte, headers map[string]string) {
	var b strings.Builder
	b.WriteString("curl -X " + method + " " + shellQuote(fullURL))
	// Only show an Authorization header when a token actually resolves — Torre's public path
	// sends none, so a bare REDACTED line would be misleading.
	if c.token != nil {
		if t, err := c.token(ctx); err == nil && t != "" {
			shown := "REDACTED"
			if c.ShowToken {
				shown = t
			}
			b.WriteString(" \\\n  -H " + shellQuote("Authorization: Bearer "+shown))
		}
	}
	for _, k := range sortedKeys(headers) {
		b.WriteString(" \\\n  -H " + shellQuote(k+": "+headers[k]))
	}
	if body != nil {
		b.WriteString(" \\\n  -d " + shellQuote(string(body)))
	}
	fmt.Fprintln(c.DryRunOut, b.String())
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Deterministic dry-run output — never map-iteration order.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
