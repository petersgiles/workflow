package flow

import "context"

// Node is the domain-facing contract each node implements.
type Node interface {
	// ID returns a stable node identifier (used for wiring).
	ID() string
	// Enter prepares node state from input.
	Enter(ctx context.Context, in Input) (State, error)
	// Process performs the node's work and may request control actions.
	Process(ctx context.Context, s State) (Result, error)
	// Exit finalizes and returns the final result/side-effects.
	Exit(ctx context.Context, s State) (FinalResult, error)
}

// Events is a small port for emitting/observing runtime events.
type Events interface {
	Info(ctx context.Context, nodeID string, payload map[string]any)
	Success(ctx context.Context, nodeID string, payload map[string]any)
	Error(ctx context.Context, nodeID string, err error, payload map[string]any)
}
