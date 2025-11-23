package db

import "context"

// QueryRequest represents an abstract query to run against a database connector.
type QueryRequest struct {
	Collection string
	Action     string // e.g., "find", "insert"
	Filter     map[string]interface{}
	Params     []interface{}
}

// DBConnector strategy interface. Implementations must be safe for concurrent use.
type DBConnector interface {
	Connect(ctx context.Context, uri string) error
	ExecuteQuery(ctx context.Context, req QueryRequest) (interface{}, error)
	Close(ctx context.Context) error
}
