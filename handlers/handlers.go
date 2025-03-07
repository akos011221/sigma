package handlers

import (
	"net/http"
	"github.com/akos011221/sigma/core"
)

func UpdateComponent(component core.ComponentInterface) core.HandlerFunc {
	return func(c *core.Context) {
		if c.Req.Method != "POST" {
			http.Error(c.Resp, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		component.Update(c)
		c.Resp.Header().Set("Content-Type", "text/plain")
		c.Resp.WriteHeader(http.StatusOK)
		c.Resp.Write([]byte("OK"))
	}
}
