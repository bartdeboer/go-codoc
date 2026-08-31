package render

import (
	"bytes"
	"testing"
)

func TestEmptyCollectionsExplainTheResult(t *testing.T) {
	var out bytes.Buffer
	Contracts(&out, nil)
	if got := out.String(); got != "No documented contracts.\n" {
		t.Fatalf("got %q", got)
	}
}
