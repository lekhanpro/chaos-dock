package safety

import (
	"sort"
	"strings"
	"sync"
)

type TargetRegistry struct {
	mu      sync.RWMutex
	targets map[string]struct{}
}

func NewTargetRegistry() *TargetRegistry {
	return &TargetRegistry{
		targets: make(map[string]struct{}),
	}
}

func (r *TargetRegistry) Mark(containerID string) {
	id := strings.TrimSpace(containerID)
	if id == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.targets[id] = struct{}{}
}

func (r *TargetRegistry) Snapshot() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.targets))
	for id := range r.targets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (r *TargetRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.targets = make(map[string]struct{})
}

