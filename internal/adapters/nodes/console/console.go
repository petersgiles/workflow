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
	data := map[string]any{"payload": in.Payload}
	return flow.State{Data: data, Input: in}, nil
}

func (c *ConsoleNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	payload := s.Input.Payload
	if p, ok := s.Data["payload"].(map[string]any); ok {
		payload = p
	}
	log.Printf("console node=%s payload=%v", c.id, payload)
	return flow.Result{Data: payload, Signal: flow.ControlSignal{}}, nil
}

func (c *ConsoleNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
