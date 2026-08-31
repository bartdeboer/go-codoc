package verify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/bartdeboer/go-codoc/internal/load"
)

const documentedTimeout = 2 * time.Minute

type DocumentedResult struct {
	Passed bool
	Tests  map[string]TestResult
	Output string
}
type TestResult struct {
	Status string
	Output string
}
type testEvent struct {
	Action string
	Test   string
	Output string
}

// RunDocumented executes exactly the documented tests once in one bounded invocation.
func RunDocumented(ctx context.Context, pkg *load.Package, testNames []string) DocumentedResult {
	ctx, cancel := context.WithTimeout(ctx, documentedTimeout)
	defer cancel()
	alternatives := make([]string, len(testNames))
	allowed := make(map[string]bool, len(testNames))
	for i, name := range testNames {
		alternatives[i] = regexp.QuoteMeta(name)
		allowed[name] = true
	}
	pattern := "^(?:" + strings.Join(alternatives, "|") + ")$"
	cmd := exec.CommandContext(ctx, "go", "test", "-json", "-count=1", "-run", pattern, ".")
	cmd.Dir = pkg.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := DocumentedResult{Passed: err == nil, Tests: map[string]TestResult{}, Output: stderr.String()}
	scanner := bufio.NewScanner(bytes.NewReader(stdout.Bytes()))
	for scanner.Scan() {
		var event testEvent
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			result.Output += scanner.Text() + "\n"
			continue
		}
		if event.Test == "" {
			if event.Action == "output" {
				result.Output += event.Output
			}
			continue
		}
		if !allowed[event.Test] || strings.Contains(event.Test, "/") {
			continue
		}
		test := result.Tests[event.Test]
		if event.Action == "output" {
			test.Output += event.Output
		}
		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			test.Status = event.Action
		}
		result.Tests[event.Test] = test
	}
	if scanner.Err() != nil {
		result.Passed = false
		result.Output += scanner.Err().Error() + "\n"
	}
	return result
}
