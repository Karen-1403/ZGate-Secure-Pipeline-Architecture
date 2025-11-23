package pipeline

import (
	"context"
)

// QueryContext carries the request state through the pipeline
type QueryContext struct {
	Ctx      context.Context
	Request  map[string]interface{}
	User     string
	HasError bool
	Error    string
	Result   interface{}
}

// Filter is a pipeline stage
type Filter interface {
	Process(q *QueryContext) *QueryContext
}

// QueryProcessingPipeline runs filters sequentially
type QueryProcessingPipeline struct {
	Filters []Filter
}

func NewQueryProcessingPipeline(filters ...Filter) *QueryProcessingPipeline {
	return &QueryProcessingPipeline{Filters: filters}
}

func (p *QueryProcessingPipeline) Process(q *QueryContext) *QueryContext {
	current := q
	for _, f := range p.Filters {
		current = f.Process(current)
		if current.HasError {
			break
		}
	}
	return current
}
