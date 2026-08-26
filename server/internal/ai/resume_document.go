package ai

import (
	"fmt"
	"slices"

	"github.com/caslus/jobmatcha/internal/model"
)

func resumeDocumentSchema() map[string]any {
	stringArray := map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
	entry := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":        map[string]any{"type": "string"},
			"organization": map[string]any{"type": "string"},
			"location":     map[string]any{"type": "string"},
			"date_range":   map[string]any{"type": "string"},
			"highlights":   stringArray,
		},
		"required":             []any{"title", "organization", "location", "date_range", "highlights"},
		"additionalProperties": false,
	}
	section := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"heading": map[string]any{"type": "string"},
			"kind":    map[string]any{"type": "string", "enum": []any{"experience", "education", "list"}},
			"entries": map[string]any{"type": "array", "items": entry},
			"items":   stringArray,
		},
		"required":             []any{"heading", "kind", "entries", "items"},
		"additionalProperties": false,
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"header": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":    map[string]any{"type": "string"},
					"contact": stringArray,
				},
				"required":             []any{"name", "contact"},
				"additionalProperties": false,
			},
			"summary":  map[string]any{"type": "string"},
			"sections": map[string]any{"type": "array", "items": section},
		},
		"required":             []any{"header", "summary", "sections"},
		"additionalProperties": false,
	}
}

// mergeTailoredDocument accepts only wording changes to editable leaves. All
// layout-defining fields are validated and retained from the source document.
func mergeTailoredDocument(source, tailored model.ResumeDocument) (model.ResumeDocument, error) {
	if source.Header.Name != tailored.Header.Name || !slices.Equal(source.Header.Contact, tailored.Header.Contact) || len(source.Sections) != len(tailored.Sections) {
		return model.ResumeDocument{}, fmt.Errorf("header or section count changed")
	}
	result := source
	result.Summary = tailored.Summary
	for index := range source.Sections {
		originalSection := source.Sections[index]
		candidateSection := tailored.Sections[index]
		if originalSection.Heading != candidateSection.Heading || originalSection.Kind != candidateSection.Kind || len(originalSection.Entries) != len(candidateSection.Entries) || len(originalSection.Items) != len(candidateSection.Items) {
			return model.ResumeDocument{}, fmt.Errorf("section %d structure changed", index)
		}
		result.Sections[index].Items = candidateSection.Items
		for entryIndex := range originalSection.Entries {
			originalEntry := originalSection.Entries[entryIndex]
			candidateEntry := candidateSection.Entries[entryIndex]
			if originalEntry.Title != candidateEntry.Title || originalEntry.Organization != candidateEntry.Organization || originalEntry.Location != candidateEntry.Location || originalEntry.DateRange != candidateEntry.DateRange || len(originalEntry.Highlights) != len(candidateEntry.Highlights) {
				return model.ResumeDocument{}, fmt.Errorf("entry %d in section %d structure changed", entryIndex, index)
			}
			result.Sections[index].Entries[entryIndex].Highlights = candidateEntry.Highlights
		}
	}
	return result, nil
}
