package caddywasm

import (
	_ "embed"
	"github.com/caddyserver/caddy/v2/caddytest"
	"testing"
)

//go:embed testdata/Caddyfile
var config string

func TestWasm(t *testing.T) {
	tester := caddytest.NewTester(t)
	tester.InitServer(config, "caddyfile")

	tester.AssertGetResponse("http://localhost:8080/v1.0/hi", 200, "Hello /v1.0/hello")
}

func init() {
	// Workaround unreleased https://github.com/caddyserver/caddy/pull/5079
	caddytest.Default.AdminPort = 2019
}
