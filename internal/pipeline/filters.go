package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/config"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/db"
)

// AuthenticationFilter verifies credentials sent in an auth message.
type AuthenticationFilter struct{}

func (f *AuthenticationFilter) Process(q *QueryContext) *QueryContext {
	// If already authenticated, skip
	if q.User != "" {
		return q
	}

	// Accept credentials either as a dedicated auth message or embedded in other messages
	// Prefer explicit user/password fields when present.
	if u, ok := q.Request["user"].(string); ok && u != "" {
		pass, _ := q.Request["password"].(string)
		if pw, ok := config.Users[u]; !ok || pw != pass {
			q.HasError = true
			q.Error = "authentication failed"
			return q
		}
		q.User = u
		return q
	}

	// If no credentials in payload, require an explicit auth message
	if t, ok := q.Request["type"].(string); !ok || t != "auth" {
		q.HasError = true
		q.Error = "authentication required"
		return q
	}

	user, _ := q.Request["user"].(string)
	pass, _ := q.Request["password"].(string)
	if pw, ok := config.Users[user]; !ok || pw != pass {
		q.HasError = true
		q.Error = "authentication failed"
		return q
	}

	q.User = user
	return q
}

// AuthorizationFilter checks that the authenticated user is allowed to access the requested collection.
type AuthorizationFilter struct{}

func (f *AuthorizationFilter) Process(q *QueryContext) *QueryContext {
	if q.HasError {
		return q
	}
	// For query messages expect a collection field
	t, _ := q.Request["type"].(string)
	if t != "query" {
		// Not a query, leave it
		return q
	}

	collection, _ := q.Request["collection"].(string)
	if collection == "" {
		q.HasError = true
		q.Error = "missing collection"
		return q
	}

	allowed := config.UserAllowedCollections[q.User]
	for _, c := range allowed {
		if c == collection {
			return q
		}
	}

	q.HasError = true
	q.Error = "access denied to collection"
	return q
}

// QueryValidationFilter performs light validation of the request payload
type QueryValidationFilter struct{}

func (f *QueryValidationFilter) Process(q *QueryContext) *QueryContext {
	if q.HasError {
		return q
	}
	// basic validation
	t, _ := q.Request["type"].(string)
	if t == "query" {
		if _, ok := q.Request["collection"].(string); !ok {
			q.HasError = true
			q.Error = "invalid collection"
		}
	}
	return q
}

// RateLimitingFilter - a very small in-memory per-user rate limiter (requests per minute)
type RateLimitingFilter struct {
	mu      sync.Mutex
	counts  map[string]int
	resetAt time.Time
}

func NewRateLimitingFilter() *RateLimitingFilter {
	return &RateLimitingFilter{counts: map[string]int{}, resetAt: time.Now().Add(time.Minute)}
}

func (f *RateLimitingFilter) Process(q *QueryContext) *QueryContext {
	if q.HasError {
		return q
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if time.Now().After(f.resetAt) {
		f.counts = map[string]int{}
		f.resetAt = time.Now().Add(time.Minute)
	}
	f.counts[q.User]++
	if f.counts[q.User] > 60 { // 60 reqs per minute
		q.HasError = true
		q.Error = "rate limit exceeded"
	}
	return q
}

// LoggingFilter logs the request (could be extended to redact sensitive info)
type LoggingFilter struct{}

func (f *LoggingFilter) Process(q *QueryContext) *QueryContext {
	log.Printf("user=%s type=%v req=%v", q.User, q.Request["type"], q.Request)
	return q
}

// ExecutionFilter runs the query against the strategy database connector
type ExecutionFilter struct {
	Connector db.DBConnector
}

func (f *ExecutionFilter) Process(q *QueryContext) *QueryContext {
	if q.HasError {
		return q
	}

	t, _ := q.Request["type"].(string)
	switch t {
	case "query":
		// build a QueryRequest for the connector
		qr := db.QueryRequest{}
		if coll, ok := q.Request["collection"].(string); ok {
			qr.Collection = coll
		}
		if act, ok := q.Request["action"].(string); ok {
			qr.Action = act
		} else {
			qr.Action = "find"
		}
		if rawFilter, ok := q.Request["filter"].(string); ok && rawFilter != "" {
			// accept filter as JSON string
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(rawFilter), &m); err == nil {
				qr.Filter = m
			}
		}

		ctx := context.Background()
		res, err := f.Connector.ExecuteQuery(ctx, qr)
		if err != nil {
			q.HasError = true
			q.Error = fmt.Sprintf("execution error: %v", err)
			return q
		}
		q.Result = res
	default:
		q.HasError = true
		q.Error = "unsupported message type"
	}

	return q
}

// ResponseFilter formats result into a standard envelope
type ResponseFilter struct{}

func (f *ResponseFilter) Process(q *QueryContext) *QueryContext {
	if q.HasError {
		q.Result = map[string]interface{}{"status": "error", "error": q.Error}
	} else {
		q.Result = map[string]interface{}{"status": "success", "data": q.Result}
	}
	return q
}
