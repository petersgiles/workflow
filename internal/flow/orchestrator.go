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

// Extended registry contract: allow querying configured transitions for a node.
type TransitionRegistry interface {
	Registry
	Transitions(id string) (map[string]string, bool)
}

// Orchestrator executes a linear flow defined as an ordered list of node IDs.
// This is a minimal implementation; extend to support graphs/branches.
type Orchestrator struct {
	registry TransitionRegistry
	events   Events
}

func NewOrchestrator(r TransitionRegistry, e Events) *Orchestrator {
	return &Orchestrator{registry: r, events: e}
}

// (Orchestrator does not expose last result via context in the default
// simple executor. It passes the result payload as the input to the next node.)

// Run executes nodes in order. The orchestrator follows configured
// `success`/`fail` transitions declared in the runtime. Retry and Abort are supported.
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

		// Handle process-level errors as a fail path
		if err != nil {
			o.events.Error(ctx, id, err, nil)
			// consult configured transitions for this node
			if tmap, ok := o.registry.Transitions(id); ok {
				if nxt, ok2 := tmap["fail"]; ok2 {
					j := indexOf(nodeIDs, nxt)
					if j == -1 {
						return fr, fmt.Errorf("fail transition %s not in flow", nxt)
					}
					// pass original input payload to failure handler
					currentInput = state.Input
					i = j
					continue
				}
			}
			// no fail transition configured: surface the error
			return fr, err
		}

		// observe success
		o.events.Success(ctx, id, map[string]any{"result": res.Data})

		// handle control signals (Abort/Retry)
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

		// On success, consult configured transitions for this node
		if tmap, ok := o.registry.Transitions(id); ok {
			if nxt, ok2 := tmap["success"]; ok2 {
				j := indexOf(nodeIDs, nxt)
				if j == -1 {
					return fr, fmt.Errorf("success transition %s not in flow", nxt)
				}
				currentInput = Input{Payload: res.Data}
				i = j
				continue
			}
		}

		// No transition configured for this outcome: end the flow and return result
		fr.Data = res.Data
		return fr, nil
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
