package console

import (
	"context"
	"log"

	"local.com/internal/flow"
)

type ConsoleNode struct {
	id string
}

func New(id string) *ConsoleNode { return &ConsoleNode{id: id} }

func (c *ConsoleNode) ID() string { return c.id }

func (c *ConsoleNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (c *ConsoleNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	log.Printf("console node=%s payload=%v", c.id, s.Input.Payload)
	return flow.Result{Data: s.Input.Payload, Signal: flow.ControlSignal{}}, nil
}

func (c *ConsoleNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
