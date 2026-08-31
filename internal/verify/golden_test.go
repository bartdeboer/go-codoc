package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartdeboer/codoc/internal/load"
)

func TestRunGoldenReportsBuildFailureWithoutTestEvent(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module example.com/broken\n\ngo 1.24.0\n")
	write(t, filepath.Join(dir, "broken.go"), "package broken\nfunc Broken( {\n")
	pkg := &load.Package{Dir: dir}
	got := RunGolden(context.Background(), pkg)
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
