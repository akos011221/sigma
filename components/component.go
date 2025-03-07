package components

import (
	"github.com/akos011221/sigma/core"
	"html/template"
	"strings"
	"sync"
)

// Component represents a reusable UI element with state
// and a template. It holds data and behavior.
type Component struct {
	// name uniquely identifies this component
	// (e.g., "counter").
	name string

	// state is a generic map holding the component's
	// data (e.g., {"Count": 0}).
	// map[string]interface{} allows any value type.
	state map[string]interface{}

	// template is the HTML string with placeholders
	// (e.g., "{{.Count}}").
	template string

	// mu protects state from concurrent access.
	// Since updates and renders can happen in
	// different goroutines, this is critical.
	mu sync.Mutex

	// onUpdate is an optional callback for handling
	// events (e.g., button clicks).
	onUpdate func(*Component, *core.Context)
}

// NewComponent creates a new component instance.
// It's a constructor that initializes the struct.
func NewComponent(name, tmpl string, initialState map[string]interface{}, onUpdate func(*Component, *core.Context)) *Component {
	// If initialState is nil, create an empty map
	// to avoid nil pointer dereference.
	if initialState == nil {
		initialState = make(map[string]interface{})
	}

	// Ensure the component's name is in its state.
	// This is a convention for Sigma framework
	// to look up components by name.
	initialState["name"] = name

	// Return a pointer to a new Component struct.
	return &Component{
		name:     name,
		state:    initialState,
		template: tmpl,
		onUpdate: onUpdate,
	}
}

// Render generates HTML from the component's template
// and state. It uses Go's template engine to merge
// state into the template.
func (c *Component) Render() (string, error) {
	c.mu.Lock()         // Lock to safely read state
	defer c.mu.Unlock() // Unlock when done

	// Parse the template string into a *template.Template object.
	// template.Must wraps Parse() and panics on error. It assumes
	// valid templates for now.
	tmpl := template.Must(template.New(c.name).Parse(c.template))

	// strings.Builder is an efficient way to build strings
	// incrementally. It's a buffer that avoid repeated
	// string allocations.
	var buf strings.Builder

	// Execute the template with the component's state,
	// writing to the buffer. This replaces placeholders
	// like {{.Count}} with actual values.
	err := tmpl.Execute(&buf, c.state)
	if err != nil {
		return "", err // e.g., invalid state
	}

	// Return the final HTML string.
	return buf.String(), nil
}

// Update applies an event to the component (e.g., incrementing
// a counter). It calls the onUpdate callback if it exists.
func (c *Component) Update(ctx *core.Context) {
	c.mu.Lock() // Lock to safely modify state
	defer c.mu.Unlock()

	// Check if an update handler exists.
	if c.onUpdate != nil {
		// Call the callback, passing the
		// component and context.
		c.onUpdate(c, ctx)
	}
}

// State returns a read-only view of the component's state.
// It's a simple map accessor with locking.
func (c *Component) State() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// SetState updates a specific key in the state map.
// With this method it is possible to modify state 
// from outside the package.
func (c *Component) SetState(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state[key] = value
}
