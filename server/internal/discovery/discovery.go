package discovery

import (
	"context"
	"sort"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/scanner"
)

// The initial careers page often routes to several companies in a group. Keep
// enough room to inspect those immediate links without turning discovery into
// an open-ended crawl.
const defaultMaxPages = 8
const defaultMaxLinks = 100

type Options struct{ MaxPages, MaxLinks int }

type Candidate struct {
	Identity         model.BoardIdentity
	EvidenceURLs     []string
	ValidationStatus string
	ValidationError  string
}

type Result struct {
	Candidates             []Candidate
	EmployerNameSuggestion string
	Incomplete             bool
}

type Service struct {
	client   *HTTPClient
	registry *scanner.Registry
	options  Options
}

func NewService(client *HTTPClient, registry *scanner.Registry, options Options) *Service {
	if options.MaxPages <= 0 {
		options.MaxPages = defaultMaxPages
	}
	if options.MaxLinks <= 0 {
		options.MaxLinks = defaultMaxLinks
	}
	return &Service{client: client, registry: registry, options: options}
}

// Discover performs an initial fetch plus a bounded one-hop traversal of only
// relevant careers links. Known boards are classified rather than fetched.
func (s *Service) Discover(ctx context.Context, raw string) (*Result, error) {
	first, err := s.client.Fetch(ctx, raw)
	if err != nil {
		return nil, err
	}
	result := &Result{EmployerNameSuggestion: EmployerNameSuggestion(string(first.Body))}
	seenBoards := map[string]*Candidate{}
	seenPages := map[string]struct{}{first.URL.String(): {}}
	queue := []ExtractedURL{{URL: first.URL}}
	pages := 0
	for len(queue) > 0 {
		if pages >= s.options.MaxPages {
			result.Incomplete = true
			break
		}
		pageLink := queue[0]
		queue = queue[1:]
		var page *Page
		if pages == 0 {
			page = first
		} else {
			page, err = s.client.Fetch(ctx, pageLink.URL.String())
			if err != nil {
				result.Incomplete = true
				continue
			}
		}
		pages++
		if !IsHTML(page.ContentType) {
			continue
		}
		for _, link := range ExtractURLs(page.URL, string(page.Body), s.options.MaxLinks) {
			if identity, ok := s.registry.Recognize(link.URL.String()); ok {
				key := identity.Provider + ":" + identity.BoardIdentifier
				candidate := seenBoards[key]
				if candidate == nil {
					candidate = &Candidate{Identity: identity}
					seenBoards[key] = candidate
				}
				candidate.EvidenceURLs = appendUnique(candidate.EvidenceURLs, page.URL.String())
				continue
			}
			// Discovery is deliberately limited to the original careers page and
			// its immediate relevant links. We still extract known boards from
			// every fetched page, but do not follow another layer of links.
			if pages != 1 {
				continue
			}
			if !IsRelevantCareerLink(link) {
				continue
			}
			key := link.URL.String()
			if _, ok := seenPages[key]; ok {
				continue
			}
			seenPages[key] = struct{}{}
			if len(queue)+pages >= s.options.MaxPages {
				result.Incomplete = true
				continue
			}
			queue = append(queue, link)
		}
	}
	for _, candidate := range seenBoards {
		if !s.registry.Has(candidate.Identity.Provider) {
			candidate.ValidationStatus = "unsupported"
		} else if err := s.registry.Validate(ctx, candidate.Identity); err != nil {
			candidate.ValidationStatus = "invalid"
			candidate.ValidationError = err.Error()
		} else {
			candidate.ValidationStatus = "ready"
		}
		sort.Strings(candidate.EvidenceURLs)
		result.Candidates = append(result.Candidates, *candidate)
	}
	sort.Slice(result.Candidates, func(i, j int) bool {
		return result.Candidates[i].Identity.CanonicalURL < result.Candidates[j].Identity.CanonicalURL
	})
	return result, nil
}

func appendUnique(values []string, value string) []string {
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}
