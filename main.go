package main

import (
	"github.com/akos011221/sigma/components"
	"github.com/akos011221/sigma/core"
	"github.com/akos011221/sigma/handlers"
	"github.com/akos011221/sigma/realtime"
	"net/http"
)

func main() {
	app := core.New()

	counter := components.NewComponent(
		"counter",
		`<div id="counter">Count: {{.Count}}</div>`,
		map[string]interface{}{"Count": 0},
		func(c *components.Component, ctx *core.Context) {
			currentCount := c.State()["Count"].(int)
			c.SetState("Count", currentCount+1)
		},
	)
	app.RegisterComponent(counter)

	app.Handle("GET", "/", func(c *core.Context) {
		initialHTML, err := counter.Render()
		if err != nil {
			http.Error(c.Resp, "Failed to render component", http.StatusInternalServerError)
			return
		}
		html := `
		<html>
		<body>
			` + initialHTML + `
			<form id="increment-form" method="POST" action="/update/counter">
				<button type="submit">Increment</button>
			</form>
			<script>
				const form = document.getElementById("increment-form");
				form.addEventListener("submit", (e) => {
					e.preventDefault();
					fetch("/update/counter", { method: "POST" });
				});
				const evtSource = new EventSource("/sse/counter");
				evtSource.onmessage = (e) => {
					document.getElementById("counter").innerHTML = e.data;
				};
			</script>
		</body>
		</html>`
		c.Resp.Write([]byte(html))
	})

	app.Handle("GET", "/sse/counter", realtime.SSEHandler(counter))
	app.Handle("POST", "/update/counter", handlers.UpdateComponent(counter))

	http.ListenAndServe(":8080", app)
}
