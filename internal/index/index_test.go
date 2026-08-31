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
	if len(doc.Workflows) != 1 || doc.Workflows[0].ID != "get-thread" {
		t.Fatalf("workflows = %#v", doc.Workflows)
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
