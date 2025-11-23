package mvc

import (
	"context"

	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/config"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/db"
	"github.com/Karen-1403/ZGate-Secure-Pipeline-Architecture/internal/pipeline"
)

// Controller orchestrates the pipeline and the models/views
type Controller struct {
	pipeline *pipeline.QueryProcessingPipeline
}

// NewController builds a controller and wires default filters including a Mongo execution connector.
func NewController(pl *pipeline.QueryProcessingPipeline) *Controller {
	// If no pipeline passed, construct one with default filters
	if pl == nil {
		// default: will be configured per request
		pl = pipeline.NewQueryProcessingPipeline()
	}
	return &Controller{pipeline: pl}
}

// RequestContext is a lightweight wrapper returned after processing to convey user
type RequestContext struct {
	User string
}

// HandleRequest accepts a decoded JSON message (map) and returns a RequestContext and response object
func (c *Controller) HandleRequest(ctx context.Context, message map[string]interface{}) (*RequestContext, interface{}) {
	// Build pipeline for this request so ExecutionFilter can get correct DB connector per-db
	// Determine requested database (if provided) or fallback to a default database name
	dbName := "test"
	if d, ok := message["database"].(string); ok && d != "" {
		dbName = d
	}

	// Create an execution filter with a Mongo connector singleton
	mc := db.GetMongoConnector(dbName)
	// connect using configured proxy URI; note: in a real deployment, the URI would be secret and stored safely
	_ = mc.Connect(ctx, config.MongoURI)

	ef := &pipeline.ExecutionFilter{Connector: mc}

	pl := pipeline.NewQueryProcessingPipeline(
		&pipeline.AuthenticationFilter{},
		&pipeline.AuthorizationFilter{},
		&pipeline.QueryValidationFilter{},
		pipeline.NewRateLimitingFilter(),
		&pipeline.LoggingFilter{},
		ef,
		&pipeline.ResponseFilter{},
	)

	qctx := &pipeline.QueryContext{Ctx: ctx, Request: message}
	out := pl.Process(qctx)
	// Convert pipeline result into view-ready object
	return &RequestContext{User: out.User}, out.Result
}
