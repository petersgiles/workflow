package flow

import (
	"fmt"
	"os"
	"strings"
)

// WriteDOT writes a Graphviz DOT representation of the flow to path.
func WriteDOT(spec *FlowSpec, path string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "digraph %q {\n", spec.ID)
	fmt.Fprintf(&b, "  rankdir=LR;\n")
	for _, n := range spec.Nodes {
		label := n.ID
		if note, ok := n.Metadata["label"].(string); ok {
			label = note
		}
		fmt.Fprintf(&b, "  %q [label=%q, shape=box];\n", n.ID, label)
		// transitions from spec.Transitions (map) if present
		for k, v := range n.Transitions {
			// show edge label when key not empty or default
			if k == "" {
				fmt.Fprintf(&b, "  %q -> %q;\n", n.ID, v)
			} else {
				fmt.Fprintf(&b, "  %q -> %q [label=%q];\n", n.ID, v, k)
			}
		}
	}
	// If no explicit transitions, draw linear order
	if !hasAnyTransitions(spec) {
		for i := 0; i+1 < len(spec.Nodes); i++ {
			fmt.Fprintf(&b, "  %q -> %q;\n", spec.Nodes[i].ID, spec.Nodes[i+1].ID)
		}
	}
	fmt.Fprintln(&b, "}")
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func hasAnyTransitions(spec *FlowSpec) bool {
	for _, n := range spec.Nodes {
		if len(n.Transitions) > 0 {
			return true
		}
	}
	return false
}
