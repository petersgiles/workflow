package noop

import (
	"context"

	"local.com/internal/flow"
)

// NoopNode is a simple example node that echoes input and immediately succeeds.
type NoopNode struct {
	id string
}

func New(id string) *NoopNode { return &NoopNode{id: id} }

func (n *NoopNode) ID() string { return n.id }

func (n *NoopNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (n *NoopNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	// echo input payload
	out := map[string]any{}
	for k, v := range s.Input.Payload {
		out[k] = v
	}
	return flow.Result{Data: out, Signal: flow.ControlSignal{}}, nil
}

func (n *NoopNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
