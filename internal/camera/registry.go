package camera

import (
	"sync"

	"github.com/GoreeCloud/goreecloud-home-security/internal/config"
)

type Public struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type Registry struct {
	mu      sync.RWMutex
	cameras []Public
}

func NewRegistry(cameras []config.Camera) *Registry {
	items := make([]Public, 0, len(cameras))
	for _, c := range cameras {
		items = append(items, Public{ID: c.ID, Name: c.Name, Enabled: c.Enabled})
	}
	return &Registry{cameras: items}
}

func (r *Registry) Snapshot() []Public {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Public(nil), r.cameras...)
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cameras)
}
