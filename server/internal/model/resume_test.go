package model

import "testing"

func TestResumeDocumentNeedsLayoutRefinement(t *testing.T) {
	document := ResumeDocument{Sections: []ResumeSection{{
		Heading: "Technical Skills",
		Kind:    "list",
		Items:   []string{"Java", "Spring", "Go", "TypeScript", "React", "AWS", "Docker", "Linux", "Grafana", "SQL"},
	}}}
	if !document.NeedsLayoutRefinement() {
		t.Fatal("ungrouped long skills list should need refinement")
	}

	document.Sections[0].Items = []string{
		"Backend: Java, Spring, Go",
		"Frontend: TypeScript, React",
		"Cloud: AWS, Docker, Linux",
	}
	if document.NeedsLayoutRefinement() {
		t.Fatal("grouped skill rows should not need refinement")
	}
}
