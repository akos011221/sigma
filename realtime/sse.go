package realtime

import (
	"fmt"
	"net/http"
	"time"

	"github.com/akos011221/sigma/core"
)

// SSEHandler creates a handler that streams component
// updates via SSE. It's a factory function returning
// a core.HandlerFunc. It sets up a persistent connection
// to push data to the client.
func SSEHandler(component core.ComponentInterface) core.HandlerFunc {
	return func(c *core.Context) {
		// Set HTTP headers for Server-Sent Events.
		// "text/event-strem" tells the browser to
		// expect SSE data.
		c.Resp.Header().Set("Content-Type", "text/event-stream")
		// "no-cache" prevents the browser from caching
		// the stream.
		c.Resp.Header().Set("Cache-Control", "no-cache")
		// "keep-alive" ensure the connection stays open.
		c.Resp.Header().Set("Connection", "keep-alive")

		// Check if the ResponseWriter supports flushing
		// (sending data immediately).
		flusher, ok := c.Resp.(http.Flusher)
		if !ok {
			// If flushing isn't supported, return an error.
			// This is rare in modern servers but ensures
			// compatibility.
			http.Error(c.Resp, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Create a ticker to simulate updates every second.
		ticker := time.NewTicker(1 * time.Second)
		// Defer stops the ticker when the function exists
		// (e.g., client disconnects).
		defer ticker.Stop()

		// Infinite loop to stream updates. This keeps
		// the HTTP connection open until the client
		// closes it.
		for {
			// Use select to handle multiple channels concurrently.
			// This is Go's way of multiplexing I/O operations.
			select {
			case <-c.Req.Context().Done():
				// This channel closes when the client disconnects
				// Context.Done() returns a channel that's closed
				// when the request's context is canceled.
				return

			case <-ticker.C:
				// This channel receives a time value very second
				// from the ticker. Render the component's current
				// state.
				html, err := component.Render()
				if err != nil {
					// If rendering fails, send an error message
					// in SSE format.
					fmt.Fprintf(c.Resp, "data: Error: %v\n\n", err)
				} else {
					// Send the HTML as an SSE event.
					// fmt.Fprintf writes to the ResponseWriter's internal
					// buffer. SSE format requires "data: " followed by the
					// payload and two newlines.
					fmt.Fprintf(c.Resp, "data: %s\n\n", html)
				}
				// Flush sends the data immediately to the client.
				// This calls an internal method to write the
				// buffer to the network.
				flusher.Flush()
			}
		}
	}

}
