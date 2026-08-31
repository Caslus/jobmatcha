package scanner

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/caslus/jobmatcha/internal/model"
)

// Provider defines the interface every scraper adapter must implement.
type Provider interface {
	// Name returns a unique identifier (e.g., "workable", "greenhouse").
	Name() string
	// Fetch retrieves roles for the given company.
	Fetch(ctx context.Context, company *model.Company) ([]*model.Role, error)
}

// URLHash generates a stable SHA256 hash for deduplication.
func URLHash(companyID uint, rawURL string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", companyID, rawURL)))
	return fmt.Sprintf("%x", h[:16])
}

// Registry maps ATS type strings to Provider instances.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// Has reports whether an adapter type is registered.
func (r *Registry) Has(name string) bool {
	_, ok := r.providers[name]
	return ok
}
