package flow

import (
	"fmt"
)

// RuntimeRegistry is a simple registry used by the orchestrator.
// It resolves node IDs to constructed Node instances.
type RuntimeRegistry struct {
	nodes map[string]Node
}

func NewRuntimeRegistry() *RuntimeRegistry { return &RuntimeRegistry{nodes: map[string]Node{}} }

func (r *RuntimeRegistry) Resolve(id string) (Node, bool) {
	n, ok := r.nodes[id]
	return n, ok
}

func (r *RuntimeRegistry) Add(id string, n Node) { r.nodes[id] = n }

// BuildRuntime builds node instances from a FlowSpec using a factory registry.
// deps is an arbitrary struct that holds shared clients (ScriptRunner, HTTPClient, etc).
func BuildRuntime(spec *FlowSpec, fr *SimpleFactoryRegistry, deps any) (*RuntimeRegistry, []string, error) {
	rt := NewRuntimeRegistry()
	order := make([]string, 0, len(spec.Nodes))

	for _, ns := range spec.Nodes {
		if ns.ID == "" {
			return nil, nil, fmt.Errorf("node with empty id in spec")
		}
		node, err := fr.Create(ns, deps)
		if err != nil {
			return nil, nil, fmt.Errorf("creating node %s: %w", ns.ID, err)
		}
		rt.Add(ns.ID, node)
		order = append(order, ns.ID)
	}

	// Determine entrypoint order: use StartNode if set, else use spec node order.
	if spec.StartNode != "" {
		// produce an ordered list starting at StartNode and following spec.Nodes order
		startIdx := -1
		for i, id := range order {
			if id == spec.StartNode {
				startIdx = i
				break
			}
		}
		if startIdx == -1 {
			return nil, nil, fmt.Errorf("start_node %s not found", spec.StartNode)
		}
		// rotate slice so start node is first
		rot := append(order[startIdx:], order[:startIdx]...)
		return rt, rot, nil
	}

	return rt, order, nil
}
