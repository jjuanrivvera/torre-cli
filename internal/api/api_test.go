package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient points both Torre hosts at one httptest server (routed by path) and
// disables retry backoff for speed.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClientWithBaseURL(srv.URL, WithHTTPClient(srv.Client()), WithMaxRetries(0))
}

func TestSearchOpportunities(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/opportunities/_search/", r.URL.Path)
		assert.Equal(t, "20", r.URL.Query().Get("size"))
		assert.Equal(t, "false", r.URL.Query().Get("aggregate"))
		body, _ := readBody(r)
		assert.Contains(t, string(body), `"skill/role"`)
		assert.Contains(t, string(body), `"experience":"potential-to-develop"`)
		_, _ = w.Write([]byte(`{"total":2,"size":20,"offset":0,"results":[{"id":"a"},{"id":"b"}]}`))
	})
	resp, raw, err := c.SearchOpportunities(t.Context(), SearchFilters{Skill: "go"}, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Results, 2)
	assert.Contains(t, string(raw), `"total":2`)
}

func TestSearchOpportunities_AllFilters(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := readBody(r)
		s := string(body)
		assert.Contains(t, s, `"remote":{"term":true}`)
		assert.Contains(t, s, `"location":{"term":"Colombia"}`)
		assert.Contains(t, s, `"organization":{"term":"Torre"}`)
		assert.Contains(t, s, `"compensation"`)
		assert.Contains(t, s, `"USD$"`)
		_, _ = w.Write([]byte(`{"total":0,"results":[]}`))
	})
	_, _, err := c.SearchOpportunities(t.Context(), SearchFilters{
		Skill: "go", Remote: true, Location: "Colombia", Organization: "Torre", Compensation: 3000,
	}, 20, 0)
	require.NoError(t, err)
}

func TestSearchOpportunities_EmptyBody(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := readBody(r)
		assert.Equal(t, "{}", strings.TrimSpace(string(body)))
		_, _ = w.Write([]byte(`{"total":0,"results":[]}`))
	})
	_, _, err := c.SearchOpportunities(t.Context(), SearchFilters{}, 20, 0)
	require.NoError(t, err)
}

func TestSearchOpportunitiesAll_Paginates(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		offset := r.URL.Query().Get("offset")
		switch offset {
		case "0":
			_, _ = w.Write([]byte(`{"total":3,"size":2,"results":[{"id":"a"},{"id":"b"}]}`))
		default:
			_, _ = w.Write([]byte(`{"total":3,"size":2,"results":[{"id":"c"}]}`))
		}
	})
	got, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, 2, 0, true)
	require.NoError(t, err)
	assert.Len(t, got, 3)
	assert.Equal(t, 2, calls)
}

// TestSearchOpportunitiesAll_DistinctPages proves that when the server honors offset and
// serves distinct windows, --limit N returns N DISTINCT ids across pages (no truncation, no
// duplicates). The fake caps each response below the requested page size to force iteration.
func TestSearchOpportunitiesAll_DistinctPages(t *testing.T) {
	const perPage, total = 3, 30
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		off, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		var ids []string
		for i := off; i < off+perPage && i < total; i++ {
			ids = append(ids, fmt.Sprintf(`{"id":"opp-%d"}`, i))
		}
		_, _ = fmt.Fprintf(w, `{"total":%d,"size":%d,"offset":%d,"results":[%s]}`,
			total, perPage, off, strings.Join(ids, ","))
	})
	got, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, perPage, 10, true)
	require.NoError(t, err)
	assert.Len(t, got, 10, "--limit 10 must return exactly 10")
	assert.Equal(t, 10, countUniqueIDs(t, got), "all 10 ids must be distinct")
}

// TestSearchOpportunitiesAll_DedupsOverlap proves the safety-net: even when the server returns
// overlapping windows (the same id in consecutive pages), the accumulated set is de-duplicated
// by .id.
func TestSearchOpportunitiesAll_DedupsOverlap(t *testing.T) {
	// Each page shares its last id with the next page's first id (a 1-item overlap).
	pages := map[string]string{
		"0": `{"total":10,"size":3,"offset":0,"results":[{"id":"a"},{"id":"b"},{"id":"c"}]}`,
		"3": `{"total":10,"size":3,"offset":3,"results":[{"id":"c"},{"id":"d"},{"id":"e"}]}`,
		"6": `{"total":10,"size":3,"offset":6,"results":[{"id":"e"},{"id":"f"}]}`,
	}
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		off := r.URL.Query().Get("offset")
		body, ok := pages[off]
		if !ok {
			body = `{"total":10,"size":3,"offset":9,"results":[]}`
		}
		_, _ = w.Write([]byte(body))
	})
	got, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, 3, 0, true)
	require.NoError(t, err)
	// a,b,c,d,e,f = 6 unique despite c and e appearing twice on the wire.
	assert.Equal(t, 6, len(got))
	assert.Equal(t, 6, countUniqueIDs(t, got), "overlapping ids must be de-duplicated")
}

// TestSearchOpportunitiesAll_OffsetIgnored_NoDuplicates reproduces the real Torre bug: the
// server ignores offset and returns the SAME page every request. --limit 100 must yield the
// page's distinct ids ONCE (not 5x duplicated), and the loop must stop instead of spinning to
// the page cap.
func TestSearchOpportunitiesAll_OffsetIgnored_NoDuplicates(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++ // offset deliberately ignored — always page 1
		var ids []string
		for i := 0; i < 20; i++ {
			ids = append(ids, fmt.Sprintf(`{"id":"job-%d"}`, i))
		}
		_, _ = fmt.Fprintf(w, `{"total":646,"size":20,"offset":0,"results":[%s]}`, strings.Join(ids, ","))
	})
	got, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, 20, 100, false)
	require.NoError(t, err)
	assert.Equal(t, 20, len(got), "an offset-ignoring server yields only its 20 distinct rows")
	assert.Equal(t, 20, countUniqueIDs(t, got), "no duplicates despite the server repeating page 1")
	assert.LessOrEqual(t, calls, 2, "loop must stop once a page adds nothing new, not spin to the cap")
}

// TestSearchOpportunitiesAll_CapsPageSize proves the requested page size never exceeds the
// server's max of 99 (size>=100 => HTTP 400 on the real endpoint).
func TestSearchOpportunitiesAll_CapsPageSize(t *testing.T) {
	var maxSeen int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		s, _ := strconv.Atoi(r.URL.Query().Get("size"))
		if s > maxSeen {
			maxSeen = s
		}
		_, _ = w.Write([]byte(`{"total":5,"size":99,"offset":0,"results":[{"id":"a"},{"id":"b"},{"id":"c"},{"id":"d"},{"id":"e"}]}`))
	})
	_, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, 20, 1000, false)
	require.NoError(t, err)
	assert.LessOrEqual(t, maxSeen, 99, "must never request size>=100")
}

// countUniqueIDs decodes each result's .id and returns the count of distinct non-empty ids.
func countUniqueIDs(t *testing.T, results []json.RawMessage) int {
	t.Helper()
	seen := map[string]struct{}{}
	for _, r := range results {
		var env struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(r, &env))
		if env.ID != "" {
			seen[env.ID] = struct{}{}
		}
	}
	return len(seen)
}

func TestSearchOpportunitiesAll_LimitCaps(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"total":100,"size":5,"results":[{"id":"a"},{"id":"b"},{"id":"c"},{"id":"d"},{"id":"e"}]}`))
	})
	got, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, 5, 3, false)
	require.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestSearchOpportunitiesAll_SinglePage(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"total":50,"size":2,"results":[{"id":"a"},{"id":"b"}]}`))
	})
	got, err := c.SearchOpportunitiesAll(t.Context(), SearchFilters{}, 2, 0, false)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, 1, calls, "single-page mode must not follow more pages")
}

func TestGetOpportunity(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/suite/opportunities/KWN4QjAd", r.URL.Path)
		_, _ = w.Write([]byte(`{"id":"KWN4QjAd","objective":"Go Engineer"}`))
	})
	body, err := c.GetOpportunity(t.Context(), "KWN4QjAd")
	require.NoError(t, err)
	assert.Contains(t, string(body), "Go Engineer")
}

func TestGenome(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/genome/bios/torrenegra", r.URL.Path)
		_, _ = w.Write([]byte(`{"person":{"name":"Alex"}}`))
	})
	body, err := c.Genome(t.Context(), "torrenegra")
	require.NoError(t, err)
	assert.Contains(t, string(body), "Alex")
}

func TestSearchPeople(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/people/_search", r.URL.Path)
		body, _ := readBody(r)
		assert.Contains(t, string(body), `"skill/role"`)
		_, _ = w.Write([]byte(`{"total":1,"results":[{"name":"x"}]}`))
	})
	resp, _, err := c.SearchPeople(t.Context(), PeopleFilters{Skill: "go", Remote: true, Location: "CO"}, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
}

func TestSearchPeopleAll(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "0" {
			_, _ = w.Write([]byte(`{"total":2,"size":1,"results":[{"id":"a"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"total":2,"size":1,"results":[{"id":"b"}]}`))
	})
	got, err := c.SearchPeopleAll(t.Context(), PeopleFilters{}, 1, 0, true)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestDo_RawHostRouting(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"path":%q}`, r.URL.Path)
	})
	status, _, body, err := c.Do(t.Context(), "search", http.MethodGet, "people/_search", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Contains(t, string(body), "/people/_search")
}

func TestAPIError_Hints(t *testing.T) {
	cases := []struct {
		status int
		body   string
		hint   string
	}{
		{404, `{"error":"Not Found"}`, "not found"},
		{429, `{}`, "rate limited"},
		{400, `{"message":"bad"}`, "filter"},
		{500, `{"meta":{"message":"boom"}}`, "server error"},
		{401, `{}`, "token"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := c.Genome(t.Context(), "x")
			require.Error(t, err)
			var apiErr *APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tc.status, apiErr.StatusCode)
			assert.Contains(t, strings.ToLower(err.Error()), tc.hint)
		})
	}
}

func TestAPIError_MessageFromErrorsArray(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"errors":[{"code":"020000","message":"gone"}]}`))
	})
	_, err := c.Genome(t.Context(), "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gone")
}

func TestDryRun_PrintsCurl_NoAuthWhenPublic(t *testing.T) {
	var buf bytes.Buffer
	c := New("", "", WithDryRun(true, &buf))
	_, err := c.Genome(t.Context(), "torrenegra")
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "curl -X GET")
	assert.Contains(t, out, "genome/bios/torrenegra")
	assert.NotContains(t, out, "Authorization", "public path must not show an auth header")
}

func TestDryRun_RedactsToken(t *testing.T) {
	var buf bytes.Buffer
	c := New("", "", WithDryRun(true, &buf), WithToken(func(context.Context) (string, error) { return "secret123", nil }))
	_, err := c.Genome(t.Context(), "x")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Bearer REDACTED")
	assert.NotContains(t, buf.String(), "secret123")
}

func TestDryRun_ShowToken(t *testing.T) {
	var buf bytes.Buffer
	c := New("", "", WithDryRun(true, &buf), WithToken(func(context.Context) (string, error) { return "secret123", nil }))
	c.ShowToken = true
	_, _, err := c.SearchOpportunities(t.Context(), SearchFilters{Skill: "go"}, 20, 0)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Bearer secret123")
}

func TestBaseURLAccessors(t *testing.T) {
	c := New("https://s.example/", "https://a.example/")
	assert.Equal(t, "https://s.example", c.SearchBaseURL())
	assert.Equal(t, "https://a.example", c.APIBaseURL())
}

func TestRetry_On500ThenSuccess(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(500)
			return
		}
		_, _ = w.Write([]byte(`{"person":{}}`))
	}))
	defer srv.Close()
	c := NewClientWithBaseURL(srv.URL, WithHTTPClient(srv.Client()), WithMaxRetries(2))
	_, err := c.Genome(t.Context(), "x") // GET is idempotent → retried
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func readBody(r *http.Request) ([]byte, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(r.Body)
	return b.Bytes(), err
}

// ensure json import is used even if a case is trimmed
var _ = json.Marshal
