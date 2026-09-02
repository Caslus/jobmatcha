package scanner

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/url"
	"strings"

	"github.com/caslus/jobmatcha/internal/model"
)

// Provider defines the interface every scraper adapter must implement.
type Provider interface {
	// Name returns a unique identifier (e.g., "workable", "greenhouse").
	Name() string
	// Fetch retrieves roles for a company's board.
	Fetch(ctx context.Context, company *model.Company, board *model.CareerBoard) ([]*model.Role, error)
}

// BoardRecognizer is an optional adapter capability for discovery. Providers
// that only support scanning deliberately need not implement it.
type BoardRecognizer interface {
	RecognizeBoard(*url.URL) (model.BoardIdentity, bool)
	ValidateBoard(context.Context, model.BoardIdentity) error
}

// URLHash generates a stable SHA256 hash for deduplication.
func URLHash(companyID uint, rawURL string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", companyID, rawURL)))
	return fmt.Sprintf("%x", h[:16])
}

// Registry maps ATS type strings to Provider instances.
type Registry struct {
	providers   map[string]Provider
	recognizers map[string]BoardRecognizer
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider), recognizers: make(map[string]BoardRecognizer)}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
	if recognizer, ok := p.(BoardRecognizer); ok {
		r.recognizers[p.Name()] = recognizer
	}
}

// RegisterRecognizer makes a known provider discoverable even if it is not
// currently fetchable by the scanner.
func (r *Registry) RegisterRecognizer(name string, recognizer BoardRecognizer) {
	r.recognizers[name] = recognizer
}

func (r *Registry) Recognize(raw string) (model.BoardIdentity, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return model.BoardIdentity{}, false
	}
	for _, recognizer := range r.recognizers {
		if identity, ok := recognizer.RecognizeBoard(u); ok {
			return identity, true
		}
	}
	return model.BoardIdentity{}, false
}

func (r *Registry) Validate(ctx context.Context, identity model.BoardIdentity) error {
	recognizer, ok := r.recognizers[strings.ToLower(identity.Provider)]
	if !ok {
		return fmt.Errorf("no recognizer for provider %q", identity.Provider)
	}
	return recognizer.ValidateBoard(ctx, identity)
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
