package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// searchPeoplePath is the people search endpoint on the search cluster.
const searchPeoplePath = "/people/_search"

// PeopleFilters are the user-facing people-search filters. A skill search requires an
// experience level (Torre 500s without one — verified in recon), so a default is applied.
type PeopleFilters struct {
	Skill      string
	Experience string
	Remote     bool
	Location   string
}

// buildPeopleQuery compiles the filters into Torre's people query body. A bare skill/role
// with only `text` is rejected by Torre, so experience defaults to "potential-to-develop".
func buildPeopleQuery(f PeopleFilters) map[string]any {
	m := map[string]any{}
	if f.Skill != "" {
		exp := f.Experience
		if exp == "" {
			exp = "potential-to-develop"
		}
		m["skill/role"] = map[string]any{"text": f.Skill, "experience": exp}
	}
	if f.Remote {
		m["remote"] = map[string]any{"term": true}
	}
	if f.Location != "" {
		m["location"] = map[string]any{"term": f.Location}
	}
	return m
}

// SearchPeople runs one page of a people search and returns the raw envelope.
func (c *Client) SearchPeople(ctx context.Context, f PeopleFilters, size, offset int) (*SearchResponse, json.RawMessage, error) {
	body, err := json.Marshal(buildPeopleQuery(f))
	if err != nil {
		return nil, nil, err
	}
	u := c.searchBase + searchPeoplePath + "?" + searchQuery(size, offset).Encode()
	raw, err := c.postJSON(ctx, u, body)
	if err != nil {
		return nil, nil, err
	}
	if raw == nil { // dry-run
		return nil, nil, nil
	}
	var resp SearchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, nil, fmt.Errorf("decode people response: %w", err)
	}
	return &resp, raw, nil
}

// SearchPeopleAll walks pages by advancing offset until `limit` results are collected.
func (c *Client) SearchPeopleAll(ctx context.Context, f PeopleFilters, size, limit int, all bool) ([]json.RawMessage, error) {
	const pageCap = 100
	if size <= 0 {
		size = 20
	}
	var out []json.RawMessage
	offset := 0
	for page := 0; ; page++ {
		if page >= pageCap {
			return out, fmt.Errorf("stopped after %d pages — narrow the query or use --limit", pageCap)
		}
		resp, _, err := c.SearchPeople(ctx, f, size, offset)
		if err != nil {
			return nil, err
		}
		if resp == nil { // dry-run
			return nil, nil
		}
		out = append(out, resp.Results...)
		if limit > 0 && len(out) >= limit {
			return out[:limit], nil
		}
		if len(resp.Results) == 0 || offset+len(resp.Results) >= resp.Total {
			return out, nil
		}
		if !all && limit == 0 {
			return out, nil
		}
		offset += len(resp.Results)
	}
}
