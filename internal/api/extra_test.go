package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_String(t *testing.T) {
	assert.Equal(t, "abc", ID("abc").String())
}

func TestWithUserAgent(t *testing.T) {
	c := New("", "", WithUserAgent("custom/1.0"))
	assert.Equal(t, "custom/1.0", c.userAgent)
	// empty keeps the default
	c2 := New("", "", WithUserAgent(""))
	assert.Equal(t, DefaultUserAgent, c2.userAgent)
}

func TestRetryAfter(t *testing.T) {
	h := http.Header{}
	h.Set("Retry-After", "2")
	assert.Equal(t, 2*time.Second, retryAfter(h))

	h.Set("Retry-After", time.Now().Add(3*time.Second).UTC().Format(http.TimeFormat))
	d := retryAfter(h)
	assert.Greater(t, d, time.Duration(0))

	assert.Equal(t, time.Duration(0), retryAfter(http.Header{}))
}

func TestIsTransient(t *testing.T) {
	assert.False(t, isTransient(context.Canceled))
	assert.False(t, isTransient(context.DeadlineExceeded))
	assert.True(t, isTransient(&net.OpError{Op: "dial", Err: errors.New("refused")}))
}

func TestBuildOpportunityQuery_Empty(t *testing.T) {
	q := buildOpportunityQuery(SearchFilters{})
	_, hasAnd := q["and"]
	assert.False(t, hasAnd, "empty filters produce an empty firehose body")
}

func TestBuildPeopleQuery_DefaultsExperience(t *testing.T) {
	q := buildPeopleQuery(PeopleFilters{Skill: "go"})
	sr, ok := q["skill/role"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "potential-to-develop", sr["experience"])
}
