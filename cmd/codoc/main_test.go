package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptions(t *testing.T) {
	got, err := options([]string{"--format", "json"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Format != "json" {
		t.Fatalf("format=%q", got.Format)
	}
}
func TestOptionsRejectsUnknownArguments(t *testing.T) {
	if _, err := options([]string{"--source"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunDefaultsToCurrentPackage(t *testing.T) {
	inFixture(t)
	var out bytes.Buffer
	if err := run(context.Background(), nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "Package fixture\n") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunContractWithoutIDListsCurrentPackageContracts(t *testing.T) {
	inFixture(t)
	var out bytes.Buffer
	if err := run(context.Background(), []string{"contract"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "get-thread/not-found") {
		t.Fatalf("output = %q", out.String())
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
