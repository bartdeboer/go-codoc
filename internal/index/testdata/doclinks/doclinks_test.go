package doclinks_test

import (
	"testing"

	target "github.com/bartdeboer/go-codoc/internal/index/testdata/doclinks"
)

// TestExternalStory uses [target.Surface], [target.Surface.Run], then [target.Client.Do].
// [target.Surface] is intentionally repeated. [target.Missing], [testing.T],
// and [target] must not create drill-down records.
//
//codoc:doc
func TestExternalStory(t *testing.T) {
	var surface target.Surface
	_ = surface
	target.Client{}.Do()
}

// TestOmittedStory also links [target.Client.Do].
//
//codoc:doc
//codoc:code omit
func TestOmittedStory(t *testing.T) { target.Client{}.Do() }
