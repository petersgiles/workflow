package flow

import (
	"fmt"
)

// NodeFactory is a constructor that builds a Node given a NodeSpec and a set of shared clients.
type NodeFactory func(spec NodeSpec, deps any) (Node, error)

// SimpleFactoryRegistry holds mappings from node type -> factory.
type SimpleFactoryRegistry struct {
	factories map[string]NodeFactory
}

// NewSimpleFactoryRegistry creates an empty registry.
func NewSimpleFactoryRegistry() *SimpleFactoryRegistry {
	return &SimpleFactoryRegistry{factories: map[string]NodeFactory{}}
}

func (r *SimpleFactoryRegistry) Register(typ string, f NodeFactory) {
	r.factories[typ] = f
}

func (r *SimpleFactoryRegistry) Create(spec NodeSpec, deps any) (Node, error) {
	f, ok := r.factories[spec.Type]
	if !ok {
		return nil, fmt.Errorf("no factory for node type: %s", spec.Type)
	}
	return f(spec, deps)
}
