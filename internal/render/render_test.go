package render

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/bartdeboer/go-codoc/internal/model"
)

func TestEmptyCollectionsExplainTheResult(t *testing.T) {
	var out bytes.Buffer
	Contracts(&out, nil)
	if got := out.String(); got != "No documented contracts.\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWorkflowShowsOutputVerificationAndSource(t *testing.T) {
	var out bytes.Buffer
	Workflow(&out, model.Workflow{ID: "story", ExpectedOutput: "done\n", Verification: "not run", Source: model.Position{File: "story_test.go", Line: 12}})
	got := out.String()
	for _, want := range []string{"Expected output:", "done", "Verification: not run", "Source: story_test.go:12"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestPackageMakesAPICapExplicit(t *testing.T) {
	var out bytes.Buffer
	symbols := make([]model.Symbol, 13)
	for i := range symbols {
		symbols[i] = model.Symbol{ID: fmt.Sprintf("Type%02d", i), DeclarationKind: "type"}
	}
	Package(&out, model.Package{Name: "many", Symbols: symbols})
	if !strings.Contains(out.String(), "API (showing 12 of 13, alphabetical):") {
		t.Fatalf("output=%q", out.String())
	}
	if strings.Contains(out.String(), "no workflows") {
		t.Fatalf("optional workflows reported as gap: %q", out.String())
	}
}

func TestNarrativeRendersContextAwareDrillDown(t *testing.T) {
	var out bytes.Buffer
	Narrative(&out, model.Narrative{Name: "contract", Passed: true, CommandDirectory: "./contract", DocumentedTests: []model.DocumentedTest{{Title: "Story", Status: "pass", RelatedSymbols: []model.RelatedSymbol{{Symbol: "CommandSurface"}, {Symbol: "CommandDefinition"}}}}})
	got := out.String()
	for _, want := range []string{"Drill down:", "go tool codoc -C ./contract symbol CommandSurface", "go tool codoc -C ./contract symbol CommandDefinition", "PASS"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}
