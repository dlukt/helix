package helix

import (
	"net/url"
	"strconv"
)

// CursorParams contains common Helix cursor pagination parameters.
type CursorParams struct {
	After  string
	Before string
	First  int
}

func addCursorParams(query url.Values, params CursorParams) {
	if params.After != "" {
		query.Set("after", params.After)
	}
	if params.Before != "" {
		query.Set("before", params.Before)
	}
	if params.First > 0 {
		query.Set("first", strconv.Itoa(params.First))
	}
}

func addRepeated(query url.Values, key string, values []string) {
	for _, value := range values {
		query.Add(key, value)
	}
}
