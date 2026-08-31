// Package load resolves a Go package and parses the files selected by go list.
package load

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Package struct {
	ImportPath string
	Name       string
	Dir        string
	Pattern    string
	WorkDir    string
	Files      []*ast.File
	FSet       *token.FileSet
}

type goListPackage struct {
	ImportPath                                   string
	Name                                         string
	Dir                                          string
	GoFiles, CgoFiles, TestGoFiles, XTestGoFiles []string
	Error                                        *struct{ Err string }
}

func PackageAt(pattern string) (*Package, error) {
	return PackageAtContext(context.Background(), "", pattern)
}

// PackageAtContext follows go list's build-tag and platform file selection.
func PackageAtContext(ctx context.Context, workDir, pattern string) (*Package, error) {
	if pattern == "" {
		pattern = "."
	}
	pattern = localPattern(workDir, pattern)
	cmd := exec.CommandContext(ctx, "go", "list", "-json", pattern)
	cmd.Dir = workDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("load %q: %s", pattern, bytes.TrimSpace(stderr.Bytes()))
	}
	var listed goListPackage
	if err := json.Unmarshal(stdout.Bytes(), &listed); err != nil {
		return nil, fmt.Errorf("decode go list output: %w", err)
	}
	if listed.Error != nil {
		return nil, fmt.Errorf("load %q: %s", pattern, listed.Error.Err)
	}
	pkg := &Package{ImportPath: listed.ImportPath, Name: listed.Name, Dir: listed.Dir, Pattern: pattern, WorkDir: workDir, FSet: token.NewFileSet()}
	files := append(append(append(listed.GoFiles, listed.CgoFiles...), listed.TestGoFiles...), listed.XTestGoFiles...)
	for _, name := range files {
		path := filepath.Join(listed.Dir, name)
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		file, err := parser.ParseFile(pkg.FSet, path, source, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		pkg.Files = append(pkg.Files, file)
	}
	return pkg, nil
}

func localPattern(workDir, pattern string) string {
	if pattern == "" || pattern == "." || strings.ContainsAny(pattern, "/\\") {
		return pattern
	}
	path := pattern
	if workDir != "" {
		path = filepath.Join(workDir, pattern)
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return "." + string(filepath.Separator) + pattern
	}
	cmd := exec.Command("go", "list", "-f", "{{.Name}}", ".")
	cmd.Dir = workDir
	name, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(name)) == pattern {
		return "."
	}
	return pattern
}
