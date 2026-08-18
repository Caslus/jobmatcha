package util

import (
	"encoding/json"
	"regexp"
)

var jsonldRe = regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)

// JobPosting represents a schema.org JobPosting JSON-LD block.
type JobPosting struct {
	Context        string          `json:"@context"`
	Type           string          `json:"@type"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	DatePosted     string          `json:"datePosted"`
	EmploymentType string          `json:"employmentType"`
	URL            string          `json:"url"`
	JobLocation    json.RawMessage `json:"jobLocation"`
}

// GraphWrapper handles @graph-wrapped JSON-LD.
type GraphWrapper struct {
	Context string       `json:"@context"`
	Graph   []JobPosting `json:"@graph"`
}

// FindJobPostingJSONLD extracts the first JobPosting JSON-LD from HTML.
func FindJobPostingJSONLD(html []byte) *JobPosting {
	matches := jsonldRe.FindAllSubmatch(html, -1)
	for _, m := range matches {
		var jp JobPosting
		if err := json.Unmarshal(m[1], &jp); err != nil {
			continue
		}
		if jp.Type == "JobPosting" {
			return &jp
		}

		// Try @graph wrapper
		var gw GraphWrapper
		if err := json.Unmarshal(m[1], &gw); err == nil {
			for _, item := range gw.Graph {
				if item.Type == "JobPosting" {
					return &item
				}
			}
		}
	}
	return nil
}