package caddywasm

import (
	"context"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	wasm "github.com/http-wasm/http-wasm-host-go/handler/nethttp"
	"net/http"
	"os"
)

func init() {
	caddy.RegisterModule(&CaddyWasm{})
	httpcaddyfile.RegisterHandlerDirective("wasm", parseCaddyfile)
}

type CaddyWasm struct {
	Path string `json:"path"`

	mw wasm.Middleware
}

func (*CaddyWasm) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.wasm",
		New: func() caddy.Module { return new(CaddyWasm) },
	}
}

func (c *CaddyWasm) Provision(ctx caddy.Context) error {
	code, err := os.ReadFile(c.Path)
	if err != nil {
		return err
	}
	mw, err := wasm.NewMiddleware(ctx, code)
	if err != nil {
		return err
	}
	c.mw = mw
	return nil
}

func (c *CaddyWasm) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	ctx := context.Background()
	h := c.mw.NewHandler(ctx, caddyHandler{next: next})
	h.ServeHTTP(w, r)
	return nil
}

type caddyHandler struct {
	next caddyhttp.Handler
}

func (c caddyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := c.next.ServeHTTP(w, r); err != nil {
		// next is called from wasm, currently panic is used to propagate errors back to the host.
		panic(err)
	}
}

func (c *CaddyWasm) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	if !d.Next() {
		return d.ArgErr()
	}
	for d.NextBlock(0) {
		key := d.Val()
		var value string
		d.Args(&value)
		switch key {
		case "path":
			c.Path = value
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	m := &CaddyWasm{}
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
