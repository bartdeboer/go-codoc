package duplicate

import "testing"

// TestFooBar is one story.
//
//codoc:doc
func TestFooBar(t *testing.T) {}

// TestFoo_Bar is another story with the same derived ID.
//
//codoc:doc
func TestFoo_Bar(t *testing.T) {}
