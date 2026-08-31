package app

import (
	"bytes"
	"encoding/json"
	"github.com/bartdeboer/codoc/internal/model"
	"testing"
)

func TestWorkflowJSONIsOneRecord(t *testing.T) {
	var out bytes.Buffer
	a := App{Out: &out}
	p := model.Package{Workflows: []model.Workflow{{Kind: "workflow", ID: "get-thread", ExampleName: "ExampleClient_GetThread"}, {Kind: "workflow", ID: "append-message", ExampleName: "ExampleAppend"}}}
	if err := a.Workflow(p, "get-thread", Options{Format: "json"}); err != nil {
		t.Fatal(err)
	}
	var got model.Workflow
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "get-thread" {
		t.Fatalf("id=%q", got.ID)
	}
	if bytes.Contains(out.Bytes(), []byte("append-message")) {
		t.Fatal("unrelated record returned")
	}
}
func TestWorkflowNotFound(t *testing.T) {
	err := (App{Out: &bytes.Buffer{}}).Workflow(model.Package{}, "missing", Options{})
	if err == nil {
		t.Fatal("expected error")
	}
}
