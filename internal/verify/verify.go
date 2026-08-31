// Package verify explicitly runs package tests for documentation freshness checks.
package verify

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"

	"github.com/bartdeboer/codoc/internal/load"
	"github.com/bartdeboer/codoc/internal/model"
)

func Run(ctx context.Context, pkg *load.Package, kind, target, testName string) model.Verification {
	args := []string{"test", "."}
	if testName != "" {
		args = append(args, "-run", "^"+regexp.QuoteMeta(testName)+"$")
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = pkg.Dir
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	if target == "" {
		target = pkg.ImportPath
	}
	return model.Verification{Kind: kind, Target: target, Passed: err == nil, Output: output.String()}
}
