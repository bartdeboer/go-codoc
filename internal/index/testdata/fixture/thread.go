package fixture

import "context"

// ErrNotFound reports a missing thread.
var ErrNotFound = errorString("not found")

type errorString string

func (e errorString) Error() string { return string(e) }

// Client retrieves threads.
type Client struct{}

// Thread is a conversation.
type Thread struct{ ID string }

// GetThread retrieves a thread by ID.
func (c *Client) GetThread(ctx context.Context, id string) (*Thread, error) {
	if id == "missing" {
		return nil, ErrNotFound
	}
	return &Thread{ID: id}, nil
}

// System invokes operator commands.
type System interface {
	InvokeOperatorCLI() error
}

type storySystem struct{}

func (*storySystem) TrustBrowser() {}

func TestDocumentedProductionHelper() {}
