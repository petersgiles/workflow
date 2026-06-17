package flow

import (
	"context"
	"time"
)

// NoopEvents is a simple Events implementation that does nothing.
type NoopEvents struct{}

func (NoopEvents) Info(ctx context.Context, nodeID string, payload map[string]any)             {}
func (NoopEvents) Success(ctx context.Context, nodeID string, payload map[string]any)          {}
func (NoopEvents) Error(ctx context.Context, nodeID string, err error, payload map[string]any) {}

type Input struct {
	Payload map[string]any
}

type State struct {
	// Node-specific opaque state
	Data  map[string]any
	Input Input
}

type ControlSignal struct {
	Next  string // next node ID (empty = implicit)
	Retry bool
	Abort bool
	Delay time.Duration
}

type Result struct {
	Data   map[string]any
	Signal ControlSignal
}

type FinalResult struct {
	Data   map[string]any
	Errors []string
}
