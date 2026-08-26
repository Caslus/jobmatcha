package ai

import (
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
)

func TestMergeTailoredDocumentPreservesStructure(t *testing.T) {
	source := model.ResumeDocument{
		Header:  model.ResumeHeader{Name: "Taylor", Contact: []string{"taylor@example.com"}},
		Summary: "Build APIs.",
		Sections: []model.ResumeSection{{
			Heading: "Experience", Kind: "experience",
			Entries: []model.ResumeEntry{{Title: "Engineer", Organization: "Acme", DateRange: "2024 - Present", Highlights: []string{"Built APIs."}}},
		}},
	}
	candidate := source
	candidate.Header.Contact = append([]string(nil), source.Header.Contact...)
	candidate.Sections = append([]model.ResumeSection(nil), source.Sections...)
	candidate.Sections[0].Entries = append([]model.ResumeEntry(nil), source.Sections[0].Entries...)
	candidate.Sections[0].Entries[0].Highlights = append([]string(nil), source.Sections[0].Entries[0].Highlights...)
	candidate.Summary = "Build reliable APIs."
	candidate.Sections[0].Entries[0].Highlights[0] = "Built reliable APIs."

	got, err := mergeTailoredDocument(source, candidate)
	if err != nil {
		t.Fatalf("mergeTailoredDocument() error = %v", err)
	}
	if got.Sections[0].Entries[0].Highlights[0] != "Built reliable APIs." {
		t.Fatalf("highlight = %q, want tailored text", got.Sections[0].Entries[0].Highlights[0])
	}

	candidate.Sections[0].Entries[0].Title = "Senior Engineer"
	if _, err := mergeTailoredDocument(source, candidate); err == nil {
		t.Fatal("mergeTailoredDocument() accepted a structural change")
	}
}
