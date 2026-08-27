package ai

import (
	"testing"

	"github.com/OpenRouterTeam/go-sdk/models/components"
)

func TestParseJSONResultAcceptsFencedAndSurroundingText(t *testing.T) {
	for _, input := range []string{
		"```json\n{\"document\":{\"header\":{\"name\":\"Ada\"}}}\n```",
		"Here is the result: {\"document\":{\"header\":{\"name\":\"Ada\"}}} thanks",
	} {
		result, err := parseJSONResult(input)
		if err != nil || result.Document.Header.Name != "Ada" {
			t.Fatalf("parseJSONResult(%q) = %#v, %v", input, result, err)
		}
	}
	if _, err := parseJSONResult("not json"); err == nil {
		t.Fatal("invalid JSON was accepted")
	}
}

func TestParseTailorAndExtractAssistantText(t *testing.T) {
	result, err := parseTailorResumeResultJSON("```json\n{\"document\":{\"content\":\"tailored\"}}\n```")
	if err != nil || result.Document.Content != "tailored" {
		t.Fatalf("tailor parse = %#v, %v", result, err)
	}
	text := " hello "
	if got := extractAssistantText(components.ChatAssistantMessageContent{Str: &text}); got != "hello" {
		t.Fatalf("extract string = %q", got)
	}
}
