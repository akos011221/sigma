package core

import (
	"net/http"
	"sync"
)

// Sigma is the central struct of the framework, managing routes and components.
// It acts as the "core" that ties everything together.
type Sigma struct {
	// routes maps HTTP methods (e.g., "GET", "POST") to a map of paths
	// (e.g., "/home"). This nested map structure allows method-specific
	// routing: routes["GET"]["/home"]
	routes map[string]map[string]HandlerFunc

	// components maps components names (e.g., "counter") to their
	// implementations. This registry lets us look up components
	// by name when rendering or updating.
	components map[string]ComponentInterface

	// mu is to prevent race conditions. Since Go's HTTP server runs
	// in a seperate goroutine, multiple goroutines could modify
	// routes or components concurrently without this.
	mu sync.Mutex
}

// HandlerFunc takes a *Context and handles an HTTP request. It is a custom
// defined HandlerFunc that works with Sigma's Context struct, giving more
// control.
type HandlerFunc func(*Context)

// Context is a struct that bundles request and response data for handlers.
// It acts as a "carrier" of information through the request lifecycle.
// Under the hood, it's just a struct with pointers to the original HTTP
// request/response objects.
type Context struct {
	// Req is a pointer to the original HTTP request from net/http.
	// It contains method, URL, headers, body etc., sent by the client.
	Req *http.Request

	// Resp is an interface for writing response, also implemented by
	// net/http. It provides methods like Write(), WriteHeader(), and
	// Header() to send data back to the client.
	Resp http.ResponseWriter

	// Params stores route parameters (e.g., "id" from "/users/:id").
	// It's a simple key-value map, populated by the router when
	// matching dynamic paths.
	Params map[string]string
}

// ComponentInterface defines the contract all components must follow.
// Any type that implements these three methods can be stored as
// a ComponentInterface.
type ComponentInterface interface {
	// Render returns the component's HTML as a string.
	Render() (string, error)

	// Update modifies the component's state based on a request.
	Update(*Context)

	// State returns the current state as a map.
	State() map[string]interface{}
}

// New creates a new Sigma instance.
func New() *Sigma {
	return &Sigma{
		// Intialize routes as a nested map. Method->Path->Handler
		routes: make(map[string]map[string]HandlerFunc),

		// Initialize components to store registered components.
		components: make(map[string]ComponentInterface),
	}
}

// Handle registers a route for a specific HTTP method and path.
// It updates the routes map in a thread-safe way.
func (s *Sigma) Handle(method, path string, handler HandlerFunc) {
	// Lock the mutex to prevent concurrent writes to routes.
	// Required, so multiple goroutines don't corrupt the map.
	s.mu.Lock()
	// Unlock when the function exists.
	defer s.mu.Unlock()

	// If no map exists for this method (e.g., "GET), create one.
	// This lazy initialization avoids pre-allocating for unused
	// methods.
	if s.routes[method] == nil {
		s.routes[method] = make(map[string]HandlerFunc)
	}

	// Assign the handler to the path for this method.
	// The handler is just a function pointer stored
	// in the map.
	s.routes[method][path] = handler
}

// RegisterComponent adds a component to the registry.
func (s *Sigma) RegisterComponent(c ComponentInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use the component's "name" from its state as the key.
	// This assumes every component has a "name" in its
	// state map.
	s.components[c.State()["name"].(string)] = c
}

// ServeHTTP makes Sigma implement the http.Handler interface.
// Under the hood, this is called by net/http for every incoming 
// request. It's the entry point where requests are routed to 
// handlers.
func (s *Sigma) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Look up routes for the request's method (e.g., "GET").
	// methodRoutes is a map of paths to handlers, or nil if
	// the method isn't supported
	methodRoutes, ok := s.routes[r.Method]
	if !ok {
		// If no routes exist for this method, return a 405 error.
		// http.Error writes the status code and message to w.
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Look up the handler for the exact path (e.g., "/home").
	// handler is a HandlerFunc (Sigma's custom function type),
	// or nil if not found.
	handler, ok := methodRoutes[r.URL.Path]
	if !ok {
		// If no handler matches the path, return a 404 error.
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Create a new Context for this request.
	ctx := &Context{
		Req:	r,	// Pass the original request
		Resp:	w,	// Pass the original response writer
		Params: make(map[string]string), // Initialize an empty params map
	}

	// Call the handler with the context.
	// It invokes the function stored in the handler variable.
	// Since HandlerFunc is a type alias for func(*Context),
	// it's just a function call.
	handler(ctx)
}
