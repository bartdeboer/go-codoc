// Package load resolves a local Go package and parses its source and test files.
package load

import (
	"bytes"
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
	Files      []*ast.File
	FSet       *token.FileSet
	Sources    map[string][]byte
}

type goListPackage struct {
	ImportPath   string
	Name         string
	Dir          string
	GoFiles      []string
	CgoFiles     []string
	TestGoFiles  []string
	XTestGoFiles []string
	Error        *struct{ Err string }
}

func PackageAt(pattern string) (*Package, error) {
	pattern = localPattern(pattern)
	cmd := exec.Command("go", "list", "-json", pattern)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
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

	pkg := &Package{ImportPath: listed.ImportPath, Name: listed.Name, Dir: listed.Dir, FSet: token.NewFileSet(), Sources: map[string][]byte{}}
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
		pkg.Sources[path] = source
	}
	return pkg, nil
}

// localPattern makes package names behave like paths when that is unambiguous.
// Go intentionally treats "contract" as an import path, while CLI users commonly
// mean ./contract or the package in their current directory.
func localPattern(pattern string) string {
	if pattern == "" || pattern == "." || strings.ContainsAny(pattern, "/\\") {
		return pattern
	}
	if info, err := os.Stat(pattern); err == nil && info.IsDir() {
		return "." + string(filepath.Separator) + pattern
	}
	cmd := exec.Command("go", "list", "-f", "{{.Name}}", ".")
	name, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(name)) == pattern {
		return "."
	}
	return pattern
}
