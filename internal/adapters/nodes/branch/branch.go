package branch

import (
	"context"

	"local.com/internal/flow"
)

type BranchNode struct {
	id    string
	expr  string
	ev    flow.ExprEvaluator
	trans map[string]string // optional transitions mapping, e.g., "true": "nextID", "false":"altID"
}

func New(id string, expr string, ev flow.ExprEvaluator, trans map[string]string) *BranchNode {
	return &BranchNode{id: id, expr: expr, ev: ev, trans: trans}
}

func (b *BranchNode) ID() string { return b.id }

func (b *BranchNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (b *BranchNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	ok, err := b.ev.EvalBool(ctx, b.expr, s.Input.Payload)
	if err != nil {
		return flow.Result{}, err
	}
	sig := flow.ControlSignal{}
	// pick next based on transitions if provided
	if b.trans != nil {
		if ok {
			if nxt, found := b.trans["true"]; found {
				sig.Next = nxt
			}
		} else {
			if nxt, found := b.trans["false"]; found {
				sig.Next = nxt
			}
		}
	} else {
		// default: encode result into payload for orchestrator to handle transitions
		if ok {
			s.Data["branch"] = "true"
		} else {
			s.Data["branch"] = "false"
		}
	}
	return flow.Result{Data: s.Data, Signal: sig}, nil
}

func (b *BranchNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
