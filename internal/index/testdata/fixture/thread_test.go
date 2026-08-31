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
