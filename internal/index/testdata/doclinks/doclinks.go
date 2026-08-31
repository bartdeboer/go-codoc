// Package doclinks provides symbols used by documented test links.
package doclinks

// Surface is an architectural boundary.
type Surface interface{ Run() }

// Client invokes operations.
type Client struct{}

// Do invokes one operation.
func (Client) Do() {}
