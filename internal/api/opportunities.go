package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Opportunity is a typed slice of a search result — enough to render a table and prove the
// flexible types decode Torre's mixed id shapes. The full record is always available via
// -o json (unknown fields are ignored).
type Opportunity struct {
	ID          ID            `json:"id,omitempty"`
	Objective   string        `json:"objective,omitempty"`
	Tagline     string        `json:"tagline,omitempty"`
	Opportunity string        `json:"opportunity,omitempty"`
	Type        string        `json:"type,omitempty"`
	Status      string        `json:"status,omitempty"`
	Remote      Bool          `json:"remote,omitempty"`
	Commitment  string        `json:"commitment,omitempty"`
	Locations   StringOrSlice `json:"locations,omitempty"`
	Slug        string        `json:"slug,omitempty"`
}

// SearchResponse is the search cluster's collection envelope.
type SearchResponse struct {
	Total   int               `json:"total"`
	Size    int               `json:"size"`
	Offset  int               `json:"offset"`
	Results []json.RawMessage `json:"results"`
}

// SearchFilters are the user-facing opportunity search filters. They compile into Torre's
// `and`-array query DSL (see buildOpportunityQuery). Zero-valued fields are omitted.
type SearchFilters struct {
	Skill        string // free text matched against skill/role
	Experience   string // required skill experience level (Torre's seniority proxy)
	Remote       bool
	Location     string
	Compensation float64
	Currency     string
	Periodicity  string
	Organization string
}

// searchOpportunitiesPath is the opportunities search endpoint (trailing slash required by
// the search cluster).
const searchOpportunitiesPath = "/opportunities/_search/"

// buildOpportunityQuery compiles the filters into Torre's boolean query body:
//
//	{"and":[{"skill/role":{"text":"go","experience":"potential-to-develop"}},{"remote":{"term":true}}, ...]}
//
// A skill search requires an experience level (Torre 400s without one — verified in recon),
// so a default is applied when a skill is given.
func buildOpportunityQuery(f SearchFilters) map[string]any {
	var and []map[string]any
	if f.Skill != "" {
		exp := f.Experience
		if exp == "" {
			exp = "potential-to-develop"
		}
		and = append(and, map[string]any{"skill/role": map[string]any{"text": f.Skill, "experience": exp}})
	}
	if f.Remote {
		and = append(and, map[string]any{"remote": map[string]any{"term": true}})
	}
	if f.Location != "" {
		and = append(and, map[string]any{"location": map[string]any{"term": f.Location}})
	}
	if f.Organization != "" {
		and = append(and, map[string]any{"organization": map[string]any{"term": f.Organization}})
	}
	if f.Compensation > 0 {
		cur := f.Currency
		if cur == "" {
			cur = "USD$"
		}
		per := f.Periodicity
		if per == "" {
			per = "monthly"
		}
		and = append(and, map[string]any{"compensation": map[string]any{
			"value": f.Compensation, "currency": cur, "periodicity": per,
		}})
	}
	if len(and) == 0 {
		// An empty body returns the full firehose; Torre accepts {} for "everything".
		return map[string]any{}
	}
	return map[string]any{"and": and}
}

// searchQuery builds the size/offset/aggregate query string.
func searchQuery(size, offset int) url.Values {
	q := url.Values{}
	q.Set("size", strconv.Itoa(size))
	q.Set("offset", strconv.Itoa(offset))
	q.Set("aggregate", "false")
	return q
}

// SearchOpportunities runs one page of an opportunity search and returns the raw envelope.
func (c *Client) SearchOpportunities(ctx context.Context, f SearchFilters, size, offset int) (*SearchResponse, json.RawMessage, error) {
	body, err := json.Marshal(buildOpportunityQuery(f))
	if err != nil {
		return nil, nil, err
	}
	u := c.searchBase + searchOpportunitiesPath + "?" + searchQuery(size, offset).Encode()
	raw, err := c.postJSON(ctx, u, body)
	if err != nil {
		return nil, nil, err
	}
	if raw == nil { // dry-run
		return nil, nil, nil
	}
	var resp SearchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, nil, fmt.Errorf("decode search response: %w", err)
	}
	return &resp, raw, nil
}

// maxOpportunityPageSize is the largest `size` the opportunities _search cluster accepts;
// size>=100 is rejected with HTTP 400 "Request size too large". Verified live 2026-07-19
// (DECISIONS.md #17).
const maxOpportunityPageSize = 99

// SearchOpportunitiesAll collects up to `limit` DISTINCT opportunities (limit<=0 && !all
// means one page). Two Torre realities shape this (DECISIONS.md #17, verified live
// 2026-07-19):
//
//  1. The opportunities _search endpoint SILENTLY IGNORES the `offset` query param — every
//     request returns `offset:0`, so a naive offset loop re-fetches page 1 each iteration and
//     emits the same rows N times (the "5x duplicate" bug). The only lever that works is
//     `size`, capped at 99. So the page is sized to the demand (bounded by the cap) and, if
//     the server ever honors offset again, the loop still advances it.
//  2. As a hard safety net, results are de-duplicated by `.id` while accumulating, so the CLI
//     never emits a duplicate opportunity even if the server returns overlapping windows.
func (c *Client) SearchOpportunitiesAll(ctx context.Context, f SearchFilters, size, limit int, all bool) ([]json.RawMessage, error) {
	const pageCap = 100
	if size <= 0 {
		size = 20
	}
	// Because offset paging is a no-op on this endpoint, the only way to surface more than
	// one response's worth of results is a single larger page. Size it to the demand.
	pageSize := size
	switch {
	case all:
		pageSize = maxOpportunityPageSize
	case limit > pageSize:
		pageSize = limit
	}
	if pageSize > maxOpportunityPageSize {
		pageSize = maxOpportunityPageSize
	}

	seen := make(map[string]struct{})
	var out []json.RawMessage
	offset := 0
	for page := 0; page < pageCap; page++ {
		resp, _, err := c.SearchOpportunities(ctx, f, pageSize, offset)
		if err != nil {
			return nil, err
		}
		if resp == nil { // dry-run prints the first request only
			return nil, nil
		}
		added := 0
		for _, r := range resp.Results {
			if id := opportunityID(r); id != "" {
				if _, dup := seen[id]; dup {
					continue // dedup safety net
				}
				seen[id] = struct{}{}
			}
			out = append(out, r)
			added++
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
		// A page that adds nothing new means the server has no more DISTINCT results to
		// give (an offset-ignoring server just repeated its first page) — stop instead of
		// spinning to the page cap.
		if added == 0 || len(resp.Results) == 0 {
			return out, nil
		}
		if offset+len(resp.Results) >= resp.Total {
			return out, nil
		}
		if !all && limit == 0 {
			return out, nil // single-page mode
		}
		offset += len(resp.Results)
	}
	return out, fmt.Errorf("stopped after %d pages — narrow the query or use --limit", pageCap)
}

// opportunityID extracts an opportunity's `.id` for de-duplication. A result with no id (or
// an unparseable one) yields "" and is never treated as a duplicate.
func opportunityID(raw json.RawMessage) string {
	var env struct {
		ID ID `json:"id"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return ""
	}
	return env.ID.String()
}

// GetOpportunity fetches one opportunity's full detail from the app API.
func (c *Client) GetOpportunity(ctx context.Context, id string) (json.RawMessage, error) {
	u := c.apiBase + "/suite/opportunities/" + url.PathEscape(id)
	return c.getJSON(ctx, u, nil)
}
