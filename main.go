package main

import (
	"github.com/akos011221/sigma/components"
	"github.com/akos011221/sigma/core"
	"github.com/akos011221/sigma/handlers"
	"github.com/akos011221/sigma/realtime"
	"net/http"
	"fmt"
)

func main() {
	app := core.New()

	// Todo list component
	todoList := components.NewComponent(
		"todo-list",
		`<div id="todo-list">
			<ul>
				{{range .Todos}}
					<li><input type="checkbox" {{if .Done}}checked{{end}} data-id="{{.ID}}"> {{.Text}}</li>
				{{end}}
			</ul>
		</div>`,
		map[string]interface{}{
			"Todos": []map[string]interface{}{
				{"ID": 1, "Text": "Buy milk", "Done": false},
				{"ID": 2, "Text": "Walk dog", "Done": false}, // Fixed line, ensured proper key-value pairs
			},
		},
		func(c *components.Component, ctx *core.Context) {
			if err := ctx.Req.ParseForm(); err != nil {
				return
			}
			action := ctx.Req.FormValue("action")
			switch action {
			case "add":
				text := ctx.Req.FormValue("text")
				if text != "" {
					todos := c.State()["Todos"].([]map[string]interface{})
					newID := len(todos) + 1
					todos = append(todos, map[string]interface{}{
						"ID":   newID,
						"Text": text,
						"Done": false,
					})
					c.SetState("Todos", todos)
				}
			case "toggle":
				idStr := ctx.Req.FormValue("id")
				if idStr != "" {
					todos := c.State()["Todos"].([]map[string]interface{})
					for i, todo := range todos {
						// Convert idStr to int for comparison
						if fmt.Sprintf("%d", todo["ID"]) == idStr {
							todo["Done"] = !todo["Done"].(bool)
							todos[i] = todo
							break
						}
					}
					c.SetState("Todos", todos)
				}
			}
		},
	)
	app.RegisterComponent(todoList)

	// Home page
	app.Handle("GET", "/", func(c *core.Context) {
		initialHTML, err := todoList.Render()
		if err != nil {
			http.Error(c.Resp, "Failed to render component", http.StatusInternalServerError)
			return
		}
		html := `
		<html>
		<body>
			` + initialHTML + `
			<form id="add-todo" method="POST" action="/update/todo-list">
				<input type="text" name="text" placeholder="New todo">
				<input type="hidden" name="action" value="add">
				<button type="submit">Add</button>
			</form>
			<script>
				const addForm = document.getElementById("add-todo");
				addForm.addEventListener("submit", (e) => {
					e.preventDefault();
					fetch("/update/todo-list", {
						method: "POST",
						body: new FormData(addForm)
					});
				});
				const todoList = document.getElementById("todo-list");
				todoList.addEventListener("change", (e) => {
					if (e.target.type === "checkbox") {
						const id = e.target.getAttribute("data-id");
						fetch("/update/todo-list", {
							method: "POST",
							body: new URLSearchParams({
								action: "toggle",
								id: id
							})
						});
					}
				});
				const evtSource = new EventSource("/sse/todo-list");
				evtSource.onmessage = (e) => {
					document.getElementById("todo-list").innerHTML = e.data;
				};
			</script>
		</body>
		</html>`
		c.Resp.Write([]byte(html))
	})

	app.Handle("GET", "/sse/todo-list", realtime.SSEHandler(todoList))
	app.Handle("POST", "/update/todo-list", handlers.UpdateComponent(todoList))

	http.ListenAndServe(":8080", app)
}
