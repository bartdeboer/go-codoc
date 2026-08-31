package query

import (
	"github.com/bartdeboer/codoc/internal/model"
	"testing"
)

func TestSearchRanksWorkflowIdentifier(t *testing.T) {
	p := model.Package{Workflows: []model.Workflow{{ID: "get-thread", Summary: "Retrieve a conversation"}}, Symbols: []model.Symbol{{ID: "Thread", Doc: "A conversation"}}}
	got := Search(p, "get thread", 5)
	if len(got) == 0 || got[0].ID != "get-thread" {
		t.Fatalf("matches = %#v", got)
	}
}

func TestSearchIncludesTypeSignatures(t *testing.T) {
	p := model.Package{Symbols: []model.Symbol{{ID: "System", Signature: "type System interface { InvokeOperatorCLI() error }"}}}
	got := Search(p, "InvokeOperatorCLI", 5)
	if len(got) != 1 || got[0].ID != "System" {
		t.Fatalf("matches=%#v", got)
	}
}
