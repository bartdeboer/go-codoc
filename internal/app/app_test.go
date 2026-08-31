package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bartdeboer/go-codoc/internal/model"
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

func TestWorkflowJSONUsesEmptyArrays(t *testing.T) {
	var out bytes.Buffer
	a := App{Out: &out}
	p := model.Package{Workflows: []model.Workflow{{Kind: "workflow", ID: "example", ExampleName: "Example"}}}
	if err := a.Workflow(p, "example", Options{Format: "json"}); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out.Bytes(), []byte("null")) {
		t.Fatalf("JSON contains null: %s", out.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(`"related_symbols": []`)) {
		t.Fatalf("missing empty array: %s", out.String())
	}
}

func TestNarrativeWithoutDocumentedTestsDoesNotExecuteTests(t *testing.T) {
	var out bytes.Buffer
	a := App{Out: &out}
	if err := a.Narrative(context.Background(), nil, model.Package{Name: "leaf", DocumentedTests: []model.DocumentedTest{}}, Options{}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("No documented code paths.")) {
		t.Fatalf("output=%q", out.String())
	}
}
