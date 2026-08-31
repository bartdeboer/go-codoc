package verify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/bartdeboer/codoc/internal/load"
)

const goldenTimeout = 2 * time.Minute

type GoldenResult struct {
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

// RunGolden executes every promised golden path once in one bounded invocation.
func RunGolden(ctx context.Context, pkg *load.Package) GoldenResult {
	ctx, cancel := context.WithTimeout(ctx, goldenTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-json", "-count=1", "-run", "^TestGolden", ".")
	cmd.Dir = pkg.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := GoldenResult{Passed: err == nil, Tests: map[string]TestResult{}, Output: stderr.String()}
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
		if !strings.HasPrefix(event.Test, "TestGolden") || strings.Contains(event.Test, "/") {
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
