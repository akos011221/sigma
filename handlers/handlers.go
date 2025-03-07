package handlers

import (
	"github.com/akos011221/sigma/core"
	"net/http"
)

// UpdateComponent creates a handler to process component updates
// via POST. It's a factory function returning a core.HandlerFunc.
func UpdateComponent(component core.ComponentInterface) core.HandlerFunc {
	return func(c *core.Context) {
		// Check if the request method is POST.
		// c.Req.Method is a string set by the
		// client (e.g., "POST")
		if c.Req.Method != "POST" {
			// If not POST, return a 405 error.
			// Writes to the ResponseWriter.
			http.Error(c.Resp, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Call the component's Update method to process
		// the event. This delegates to the component's
		// logic (e.g., incrementing a counter).
		component.Update(c)
		
		// Prevent navigation for now.
		c.Resp.Header().Set("Content-Type", "text/plain")

		// Write a 200 OK status to indicate success.
		c.Resp.WriteHeader(http.StatusOK)

		// Ensure a response body to avoid hanging.
		c.Resp.Write([]byte("OK"))
	}
}
