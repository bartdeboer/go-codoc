package fixture

import (
	"context"
	"fmt"
	"testing"
)

// ExampleClient_GetThread demonstrates retrieving a known thread.
func ExampleClient_GetThread() {
	thread, _ := new(Client).GetThread(context.Background(), "thread-123")
	fmt.Println(thread.ID)
	// Output: thread-123
}

// GetThread returns ErrNotFound for a missing thread.
//
//codoc:contract get-thread/not-found
func TestGetThreadNotFound(t *testing.T) {
	_, err := new(Client).GetThread(context.Background(), "missing")
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestUnmarked(t *testing.T) { t.Log("not documentation") }

// ExampleSystem_trustedDeviceElevation demonstrates a trusted-device flow.
func ExampleSystem_trustedDeviceElevation() {
	fmt.Println("trusted")
	// Output: trusted
}

// ExampleSystem_illustrative has no expected output and is not a workflow.
func ExampleSystem_illustrative() { fmt.Println("not executed by go test") }

// TestThreadRetrieval demonstrates the package's core retrieval path.
//
//codoc:doc
func TestThreadRetrieval(t *testing.T) {
	thread, err := new(Client).GetThread(context.Background(), "thread-123")
	if err != nil {
		t.Fatal(err)
	}
	if thread.ID != "thread-123" {
		t.Fatalf("ID = %q", thread.ID)
	}
}

// TestHiddenMaintenance verifies a documented path whose implementation is intentionally omitted.
//
//codoc:doc
//codoc:code omit
func TestHiddenMaintenance(t *testing.T) {
	if ErrNotFound.Error() == "" {
		t.Fatal("missing sentinel")
	}
}
