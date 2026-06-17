package flow

import (
	"os"

	"gopkg.in/yaml.v3"
)

// FlowSpec represents a serialized flow file.
type FlowSpec struct {
	ID        string     `yaml:"id,omitempty"`
	Nodes     []NodeSpec `yaml:"nodes"`
	StartNode string     `yaml:"start_node,omitempty"` // optional entrypoint
}

// NodeSpec configures a node instance in a flow.
type NodeSpec struct {
	ID          string            `yaml:"id"`
	Type        string            `yaml:"type"`
	Config      map[string]any    `yaml:"config,omitempty"`
	Transitions map[string]string `yaml:"transitions,omitempty"` // e.g., "true": "next-node", "false":"alt-node"
	Metadata    map[string]any    `yaml:"metadata,omitempty"`
}

// LoadFlowSpec loads a YAML flow file from disk.
func LoadFlowSpec(path string) (*FlowSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec FlowSpec
	if err := yaml.Unmarshal(b, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}
