package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGlobalOptionsCanAppearAnywhere(t *testing.T) {
	g, args, err := parseGlobalOptions([]string{"symbol", "Client.GetThread", "--json", "-C", "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if g.format != "json" || g.workDir != "fixture" {
		t.Fatalf("options=%+v", g)
	}
	if strings.Join(args, " ") != "symbol Client.GetThread" {
		t.Fatalf("args=%v", args)
	}
}
func TestRunDefaultsToCurrentPackage(t *testing.T) {
	inFixture(t)
	var out bytes.Buffer
	if err := run(context.Background(), nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "Package fixture\n") {
		t.Fatalf("output=%q", out.String())
	}
	for _, want := range []string{"Documented code paths", "Thread Retrieval", "PASS"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in %q", want, out.String())
		}
	}
}
func TestCurrentPackageJSON(t *testing.T) {
	inFixture(t)
	var out bytes.Buffer
	if err := run(context.Background(), []string{"symbol", "Client.GetThread", "--json"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"id": "Client.GetThread"`) {
		t.Fatalf("output=%q", out.String())
	}
}
func TestVerifyWorkflow(t *testing.T) {
	inFixture(t)
	var out bytes.Buffer
	if err := run(context.Background(), []string{"verify", "workflow", "get-thread"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "PASS workflow get-thread") {
		t.Fatalf("output=%q", out.String())
	}
}
func inFixture(t *testing.T) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(old, "..", "..", "internal", "index", "testdata", "fixture")
	if err := os.Chdir(fixture); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}
