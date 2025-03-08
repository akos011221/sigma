package handlers

import (
	"net/http"

	"github.com/akos011221/sigma/core"
)

// UpdateComponent creates a handler to process component updates
// via POST. It's a factory function returning a core.HandlerFunc.
func UpdateComponent(component core.ComponentInterface) core.HandlerFunc {
	return func(c *core.Context) {
		// Check if the request method is a POST.
		if c.Req.Method != "POST" {
			// If not POST, return a 405 error,
			// writing to the ResponseWriter.
			http.Error(c.Resp, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Call the component's Update method to process
		// the event. This delegates to the component's
		// logic (e.g., incrementing a counter).
		component.Update(c)

		c.Resp.Header().Set("Content-Type", "text/plain")
		c.Resp.WriteHeader(http.StatusOK)
		c.Resp.Write([]byte("OK"))
	}
}
