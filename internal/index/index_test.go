package index_test

import (
	"strings"
	"testing"

	"github.com/bartdeboer/codoc/internal/index"
	"github.com/bartdeboer/codoc/internal/load"
	"github.com/bartdeboer/codoc/internal/model"
)

func TestBuildExtractsAddressableRecords(t *testing.T) {
	pkg, err := load.PackageAt("./testdata/fixture")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := index.Build(pkg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.Overview, "retrieves threads") {
		t.Fatalf("overview = %q", doc.Overview)
	}
	if len(doc.Workflows) != 2 || doc.Workflows[0].ID != "get-thread" {
		t.Fatalf("workflows = %#v", doc.Workflows)
	}

	workflow := doc.Workflows[0]
	if workflow.PrimarySymbol != "Client.GetThread" || workflow.ExpectedOutput != "thread-123\n" {
		t.Fatalf("workflow association/output = %#v", workflow)
	}
	trusted := workflowByID(doc, "trusted-device-elevation")
	if trusted == nil || trusted.PrimarySymbol != "System" {
		t.Fatalf("trusted workflow = %#v", trusted)
	}
	if symbol(doc, "storySystem.TrustBrowser") != nil {
		t.Fatal("method on unexported receiver was indexed")
	}
	system := symbol(doc, "System")
	if system == nil || !contains(system.RelatedWorkflows, "trusted-device-elevation") {
		t.Fatalf("System workflows = %#v", system)
	}

	if len(doc.Contracts) != 1 || doc.Contracts[0].ID != "get-thread/not-found" {
		t.Fatalf("contracts = %#v", doc.Contracts)
	}
	method := symbol(doc, "Client.GetThread")
	if method == nil {
		t.Fatal("Client.GetThread symbol not found")
	}
	if !strings.HasPrefix(method.Signature, "func (c *Client) GetThread") {
		t.Fatalf("signature = %q", method.Signature)
	}
	if strings.HasPrefix(method.Source.File, "/") {
		t.Fatalf("source path is absolute: %q", method.Source.File)
	}
	if got := doc.Workflows[0].RelatedContracts; len(got) != 1 || got[0] != "get-thread/not-found" {
		t.Fatalf("related contracts = %v", got)
	}
}

func TestBuildIgnoresUnmarkedTests(t *testing.T) {
	pkg, err := load.PackageAt("./testdata/fixture")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := index.Build(pkg)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range doc.Contracts {
		if c.TestName == "TestUnmarked" {
			t.Fatal("unmarked test indexed")
		}
	}
}

func symbol(doc model.Package, id string) *model.Symbol {
	for i := range doc.Symbols {
		if doc.Symbols[i].ID == id {
			return &doc.Symbols[i]
		}
	}
	return nil
}

func workflowByID(doc model.Package, id string) *model.Workflow {
	for i := range doc.Workflows {
		if doc.Workflows[i].ID == id {
			return &doc.Workflows[i]
		}
	}
	return nil
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
