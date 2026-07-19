package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError is a Torre API error with an actionable hint keyed by status.
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
	Body       []byte
}

// torreErrorBody covers the two error envelopes Torre returns: the search cluster's
// {timestamp,path,status,error,message} and the app API's {errors:[{code,message}]} /
// {meta:{message}}.
type torreErrorBody struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
	Meta      struct {
		Message string `json:"message"`
	} `json:"meta"`
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func parseAPIError(status int, body []byte, h http.Header) *APIError {
	e := &APIError{StatusCode: status, Body: body}
	var t torreErrorBody
	if json.Unmarshal(body, &t) == nil {
		switch {
		case t.Message != "":
			e.Message = t.Message
		case t.Meta.Message != "":
			e.Message = t.Meta.Message
		case len(t.Errors) > 0 && t.Errors[0].Message != "":
			e.Message = t.Errors[0].Message
		case t.Error != "":
			e.Message = t.Error
		}
		e.RequestID = t.RequestID
	}
	if e.Message == "" {
		e.Message = http.StatusText(status)
	}
	return e
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("Torre API error %d", e.StatusCode)
	msg += ": " + e.Message
	if e.RequestID != "" {
		msg += " (requestId " + e.RequestID + ")"
	}
	if hint := e.hint(); hint != "" {
		msg += "\nHint: " + hint
	}
	return msg
}

// hint maps a status to the remedy a user actually needs.
func (e *APIError) hint() string {
	switch e.StatusCode {
	case http.StatusUnauthorized:
		return "this endpoint needs a token — set one with `torre auth login` (most Torre endpoints are public and need none)"
	case http.StatusForbidden:
		return "access denied — the token lacks permission, or the resource is private"
	case http.StatusNotFound:
		return "not found — verify the opportunity id (from `torre jobs search`) or the username (as it appears in the profile URL)"
	case http.StatusTooManyRequests:
		return "rate limited by Torre — the CLI already honored Retry-After; slow down or narrow the query"
	case http.StatusBadRequest:
		return "Torre rejected the query — check the filter values (e.g. a skill search needs an --experience level)"
	case http.StatusInternalServerError:
		return "Torre server error — this often means a malformed search body; check the filters, or retry shortly"
	}
	if e.StatusCode >= 500 {
		return "Torre server error — usually transient, retry shortly"
	}
	return ""
}
