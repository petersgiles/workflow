package branch

import (
	"context"
	"fmt"
	"strings"

	"local.com/internal/flow"
)

type BranchNode struct {
	id       string
	expr     string
	ev       flow.ExprEvaluator
	trans    map[string]string // optional transitions mapping, e.g., "true": "nextID", "false":"altID"
	requires []string
}

func New(id string, expr string, ev flow.ExprEvaluator, trans map[string]string, requires []string) *BranchNode {
	return &BranchNode{id: id, expr: expr, ev: ev, trans: trans, requires: requires}
}

func (b *BranchNode) ID() string { return b.id }

func (b *BranchNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	// validate required preconditions if any
	for _, req := range b.requires {
		// req is a dot-separated path into payload, like "status" or "user.id"
		parts := strings.Split(req, ".")
		var cur any = in.Payload
		found := true
		for _, p := range parts {
			if m, ok := cur.(map[string]any); ok {
				if v, ok2 := m[p]; ok2 {
					cur = v
					continue
				}
			}
			found = false
			break
		}
		if !found {
			return flow.State{}, fmt.Errorf("branch: missing required payload key '%s'", req)
		}
	}
	data := map[string]any{"payload": in.Payload}
	return flow.State{Data: data, Input: in}, nil
}

func (b *BranchNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	payload := s.Input.Payload
	if p, okp := s.Data["payload"].(map[string]any); okp {
		payload = p
	}
	ok, err := b.ev.EvalBool(ctx, b.expr, payload)
	if err != nil {
		return flow.Result{}, err
	}

	if ok {
		// success: return result and let orchestrator follow `success` transition if configured
		return flow.Result{Data: s.Data, Signal: flow.ControlSignal{}}, nil
	}

	// failure: return error so orchestrator follows the `fail` transition
	return flow.Result{}, fmt.Errorf("branch: expression evaluated to false")
}

func (b *BranchNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
