package main

import "testing"

func TestOptions(t *testing.T) {
	got, err := options([]string{"--format", "json"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Format != "json" {
		t.Fatalf("format=%q", got.Format)
	}
}
func TestOptionsRejectsUnknownArguments(t *testing.T) {
	if _, err := options([]string{"--source"}); err == nil {
		t.Fatal("expected error")
	}
}
