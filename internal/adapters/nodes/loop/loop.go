package loop

import (
	"context"
	"sync"

	"local.com/internal/flow"
)

type LoopNode struct {
	id       string
	itemsKey string   // payload key that holds []any
	subflow  []string // node IDs to run per item (subflow)
	orchRun  func(ctx context.Context, nodeIDs []string, input flow.Input) (flow.FinalResult, error)
}

func New(id, itemsKey string, subflow []string, orchRun func(ctx context.Context, nodeIDs []string, input flow.Input) (flow.FinalResult, error)) *LoopNode {
	return &LoopNode{id: id, itemsKey: itemsKey, subflow: subflow, orchRun: orchRun}
}

func (l *LoopNode) ID() string { return l.id }

func (l *LoopNode) Enter(ctx context.Context, in flow.Input) (flow.State, error) {
	return flow.State{Data: map[string]any{}, Input: in}, nil
}

func (l *LoopNode) Process(ctx context.Context, s flow.State) (flow.Result, error) {
	itemsI, ok := s.Input.Payload[l.itemsKey]
	if !ok {
		return flow.Result{Data: map[string]any{"count": 0}}, nil
	}
	items, ok := itemsI.([]any)
	if !ok {
		return flow.Result{}, nil
	}
	var wg sync.WaitGroup
	errs := make([]string, 0)
	mu := sync.Mutex{}
	for _, it := range items {
		wg.Add(1)
		item := it
		go func() {
			defer wg.Done()
			_, err := l.orchRun(ctx, l.subflow, flow.Input{Payload: map[string]any{"item": item}})
			if err != nil {
				mu.Lock()
				errs = append(errs, err.Error())
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	data := map[string]any{"count": len(items)}
	if len(errs) > 0 {
		data["errors"] = errs
	}
	return flow.Result{Data: data, Signal: flow.ControlSignal{}}, nil
}

func (l *LoopNode) Exit(ctx context.Context, s flow.State) (flow.FinalResult, error) {
	return flow.FinalResult{Data: s.Data}, nil
}
