package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bartdeboer/codoc/internal/model"
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
