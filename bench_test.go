// SPDX-License-Identifier: MIT

package mux_test

import (
	"net/http"
	"testing"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/routertest"
)

func BenchmarkDefaultRouter(b *testing.B) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.URL.Path)); err != nil {
			panic(err)
		}
	}

	t := routertest.NewTester(mux.DefaultCall)
	t.Bench(b, http.HandlerFunc(h))
}

func BenchmarkContextRouter(b *testing.B) {
	h := ctxHandlerFunc(func(c *ctx) {
		if _, err := c.W.Write([]byte(c.R.URL.Path)); err != nil {
			panic(err)
		}
	})

	t := routertest.NewTester(contextCall)
	t.Bench(b, h)
}
