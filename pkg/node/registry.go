package node

import (
	"fmt"
	"sync"
)

// Registry holds registered node types
type Registry struct {
	mu    sync.RWMutex
	nodes map[string]Node
}

// NewRegistry creates a new node registry
func NewRegistry() *Registry {
	return &Registry{
		nodes: make(map[string]Node),
	}
}

// Register adds a node type to the registry
func (r *Registry) Register(n Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := n.Name()
	if _, exists := r.nodes[name]; exists {
		return fmt.Errorf("node type '%s' already registered", name)
	}

	r.nodes[name] = n
	return nil
}

// Get retrieves a node by type name
func (r *Registry) Get(name string) (Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n, ok := r.nodes[name]
	if !ok {
		return nil, fmt.Errorf("unknown node type '%s'", name)
	}

	return n, nil
}

// List returns all registered node type names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.nodes))
	for name := range r.nodes {
		names = append(names, name)
	}
	return names
}
