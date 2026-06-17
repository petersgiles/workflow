package flow

import (
	"fmt"
	"os"
	"strings"
)

// WriteMermaid writes a simple left-to-right flowchart.
func WriteMermaid(spec *FlowSpec, path string) error {
	var b strings.Builder
	fmt.Fprintln(&b, "flowchart LR")
	for _, n := range spec.Nodes {
		label := n.ID
		if note, ok := n.Metadata["label"].(string); ok {
			label = note
		}
		fmt.Fprintf(&b, "  %s[%q]\n", sanitizeID(n.ID), label)
		for k, v := range n.Transitions {
			if k == "" {
				fmt.Fprintf(&b, "  %s --> %s\n", sanitizeID(n.ID), sanitizeID(v))
			} else {
				fmt.Fprintf(&b, "  %s -- %q --> %s\n", sanitizeID(n.ID), k, sanitizeID(v))
			}
		}
	}
	if !hasAnyTransitions(spec) {
		for i := 0; i+1 < len(spec.Nodes); i++ {
			fmt.Fprintf(&b, "  %s --> %s\n", sanitizeID(spec.Nodes[i].ID), sanitizeID(spec.Nodes[i+1].ID))
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func sanitizeID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}
