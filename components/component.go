package components

import (
	"html/template"
	"strings"
	"sync"
	"github.com/akos011221/sigma/core"
)

type Component struct {
	name     string
	state    map[string]interface{}
	template string
	mu       sync.Mutex
	onUpdate func(*Component, *core.Context)
}

func NewComponent(name, tmpl string, initialState map[string]interface{}, onUpdate func(*Component, *core.Context)) *Component {
	if initialState == nil {
		initialState = make(map[string]interface{})
	}
	initialState["name"] = name
	return &Component{
		name:     name,
		state:    initialState,
		template: tmpl,
		onUpdate: onUpdate,
	}
}

func (c *Component) Render() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	tmpl := template.Must(template.New(c.name).Parse(c.template))
	var buf strings.Builder
	err := tmpl.Execute(&buf, c.state)
	return buf.String(), err
}

func (c *Component) Update(ctx *core.Context) {
	if c.onUpdate != nil {
		c.onUpdate(c, ctx)
	}
}

func (c *Component) State() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *Component) SetState(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state[key] = value
}
