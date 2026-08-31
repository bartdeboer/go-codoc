package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartdeboer/go-codoc/internal/load"
)

func TestRunDocumentedReportsBuildFailureWithoutTestEvent(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module example.com/broken\n\ngo 1.24.0\n")
	write(t, filepath.Join(dir, "broken.go"), "package broken\nfunc Broken( {\n")
	pkg := &load.Package{Dir: dir}
	got := RunDocumented(context.Background(), pkg, []string{"TestStory"})
	if got.Passed {
		t.Fatal("build failure reported as pass")
	}
}
func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunDocumentedRunsOnlySelectedTests(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module example.com/selected\n\ngo 1.24.0\n")
	write(t, filepath.Join(dir, "selected.go"), "package selected\n")
	write(t, filepath.Join(dir, "selected_test.go"), `package selected
import "testing"
func TestDocumented(t *testing.T) {}
func TestOrdinary(t *testing.T) { t.Fatal("must not run") }
`)
	result := RunDocumented(context.Background(), &load.Package{Dir: dir}, []string{"TestDocumented"})
	if !result.Passed || result.Tests["TestDocumented"].Status != "pass" {
		t.Fatalf("result=%#v", result)
	}
	if _, found := result.Tests["TestOrdinary"]; found {
		t.Fatal("ordinary test was observed")
	}
}
