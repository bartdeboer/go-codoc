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
