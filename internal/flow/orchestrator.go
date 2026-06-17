package flow

import (
	"context"
	"fmt"
	"time"
)

// Registry resolves node IDs to Node implementations.
type Registry interface {
	Resolve(id string) (Node, bool)
}

// Orchestrator executes a linear flow defined as an ordered list of node IDs.
// This is a minimal implementation; extend to support graphs/branches.
type Orchestrator struct {
	registry Registry
	events   Events
}

func NewOrchestrator(r Registry, e Events) *Orchestrator {
	return &Orchestrator{registry: r, events: e}
}

// Run executes nodes in order. If a node's Result.Signal.Next is set,
// orchestrator will jump to that node id (if present). Retry and Abort are supported.
func (o *Orchestrator) Run(ctx context.Context, nodeIDs []string, input Input) (FinalResult, error) {
	var fr FinalResult
	currentInput := input
	visited := make(map[string]int) // retry guard

	i := 0
	for i < len(nodeIDs) {
		id := nodeIDs[i]
		n, ok := o.registry.Resolve(id)
		if !ok {
			return fr, fmt.Errorf("node not found: %s", id)
		}

		o.events.Info(ctx, id, map[string]any{"phase": "enter"})
		state, err := n.Enter(ctx, currentInput)
		if err != nil {
			o.events.Error(ctx, id, err, nil)
			return fr, err
		}

		o.events.Info(ctx, id, map[string]any{"phase": "process"})
		res, err := n.Process(ctx, state)
		if err != nil {
			o.events.Error(ctx, id, err, nil)
			// allow nodes to signal retry via Result.Signal.Retry instead of error
			return fr, err
		}

		// observe success
		o.events.Success(ctx, id, map[string]any{"result": res.Data})

		// handle control signals
		if res.Signal.Abort {
			o.events.Info(ctx, id, map[string]any{"action": "abort"})
			return fr, ErrAborted
		}
		if res.Signal.Retry {
			visited[id]++
			if visited[id] > 3 {
				return fr, fmt.Errorf("retry limit reached for %s", id)
			}
			delay := res.Signal.Delay
			if delay > 0 {
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return fr, ctx.Err()
				}
			}
			// retry same node (do not advance i)
			continue
		}
		// build next input from result (simple assignment; adapt as needed)
		currentInput = Input{Payload: res.Data}

		// jump if Next specified
		if res.Signal.Next != "" {
			// find index of Next in nodeIDs
			j := indexOf(nodeIDs, res.Signal.Next)
			if j == -1 {
				return fr, fmt.Errorf("next node %s not in flow", res.Signal.Next)
			}
			i = j
			continue
		}
		// else proceed to next
		i++
	}

	// finalize
	fr.Data = currentInput.Payload
	return fr, nil
}

func indexOf(arr []string, v string) int {
	for i, s := range arr {
		if s == v {
			return i
		}
	}
	return -1
}
