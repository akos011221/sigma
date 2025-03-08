package main

import (
	"net/http"
	"fmt"

	"github.com/akos011221/sigma/components"
	"github.com/akos011221/sigma/core"
	"github.com/akos011221/sigma/handlers"
	"github.com/akos011221/sigma/realtime"
)

func main() {
	app := core.New()

	// Define a "status" component; our reusable UI piece.
	// - Name: "status" (just an ID for Sigma to track it)
	// - Template: HTML with a placeholder {{.Message}} for dynamic text.
	// - Initial State: A map with a "Message" key, starting as "System: All good".
	// - onUpdate: A function to handle changes when the user submits a new status.
	status := components.NewComponent(
		"status",
		`<div id="status"> {{.Message}} </div>`, // Gets replaced by state.
		map[string]interface{}{
			"Message": "System: All good", // this is the initial state
		},
		func(c *components.Component, ctx *core.Context) {
			// This runs when a POST hits /update/status (e.g., form submit).
			// Parse the form data from the HTTP request (ctx.Req is *http.Request).
			if err := ctx.Req.ParseForm(); err != nil {
				fmt.Println("debug: ParseForm failed:", err) // Debug: Form parsing issue?
				return // Bail if form parsing fails; this is for simplicity for now.
			}
			// Grab the "message" field from the form (e.g., "System: Panic!").
			newMessage := ctx.Req.FormValue("message")
			fmt.Println("debug: Form message:", newMessage) // Debug: Did we get the input?
			if newMessage != "" { // Only update if there's actual input.
				// Set the new message in the component's state.
				c.SetState("Message", newMessage)
				fmt.Println("New state:", c.State()["Message"]) // Debug: Did state update?
			} else {
				fmt.Println("Empty message received") // Debug: No input?
			}
		},
	)

	// Register the component with Sigma so it's available for routing and updates.
	app.RegisterComponent(status)

	// Define the root route (GET /)
	app.Handle("GET", "/", func(c *core.Context) {
		// Render the component's initial HTML (e.g., "<div id='status'>System: All good</div>").
		// Render() locks the state, fills the template, and returns a string.
		initialHTML, err := status.Render()
		if err != nil {
			// If rendering fails (e.g., bad template), send a 500 error.
			http.Error(c.Resp, "Failed to render component", http.StatusInternalServerError)
			return
		}
		// Build the full page HTML.
		// - Embed initialHTML in a <body>.
		// - Add a form to submit new status messages.
		// - Include JavaScript to handle form submission and SSE updates.
		html := `
		<html>
		<body>
			` + initialHTML + `
			<form id="status-form" method="POST" action="/update/status">
				<input type="text" name="message" placeholder="Set status">
				<button type="submit">Update</button>
			</form>
			<script>
				const form = document.getElementById("status-form");
				form.addEventListener("submit", (e) => {
					e.preventDefault();
					const message = form.querySelector("[name=message]").value;
					console.log("Sending message:", message);
					fetch("/update/status", {
						method: "POST",
						body: new URLSearchParams({ message: message })
					}).then(() => console.log("POST sent"));
				});
				const evtSource = new EventSource("/sse/status");
				evtSource.onmessage = (e) => {
					console.log("SSE update:", e.data);
					document.getElementById("status").innerHTML = e.data;
				};
			</script>
		</body>
		</html>`
		c.Resp.Write([]byte(html))
	})

	// Route for SSE updates (GET /sse/status).
	// realtime.SSEHandler streams the component’s rendered HTML every second.
	app.Handle("GET", "/sse/status", realtime.SSEHandler(status))

	// Route for status updates (POST /update/status).
	// handlers.UpdateComponent wraps our component’s Update logic in a handler.
	app.Handle("POST", "/update/status", handlers.UpdateComponent(status))

	http.ListenAndServe(":8080", app)
}
