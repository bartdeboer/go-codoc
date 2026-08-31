package load

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalPatternTreatsExistingDirectoryAsRelative(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "contract"), 0o755); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if got := localPattern(context.Background(), "", "contract"); got != "./contract" {
		t.Fatalf("got %q", got)
	}
}

func TestLocalPatternPreservesImportPath(t *testing.T) {
	if got := localPattern(context.Background(), "", "example.com/project/contract"); got != "example.com/project/contract" {
		t.Fatalf("got %q", got)
	}
}

func TestLocalPatternRecognizesCurrentPackageName(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(old, "..", "index", "testdata", "fixture")
	if err := os.Chdir(fixture); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if got := localPattern(context.Background(), "", "fixture"); got != "." {
		t.Fatalf("got %q", got)
	}
}
