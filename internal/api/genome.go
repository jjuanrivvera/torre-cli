package api

import (
	"context"
	"encoding/json"
	"net/url"
)

// Genome fetches a person's public genome/bio from the app API:
// GET https://torre.ai/api/genome/bios/<username>. The response is a large object
// (person, strengths, experiences, education, …) rendered raw so `--jq`/`-o json` can slice it.
func (c *Client) Genome(ctx context.Context, username string) (json.RawMessage, error) {
	u := c.apiBase + "/genome/bios/" + url.PathEscape(username)
	return c.getJSON(ctx, u, nil)
}
