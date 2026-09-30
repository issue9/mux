// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package ctx

import (
	"testing"

	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/routertest"
)

func BenchmarkRouter(b *testing.B) {
	h := HandlerFunc(func(c *CTX) {
		if _, err := c.W.Write([]byte(c.R.URL.Path)); err != nil {
			panic(err)
		}
	})

	t := routertest.NewTester[Handler](func(o ...mux.Option) *routertest.TestRouter[Handler] {
		router := mux.NewRouter[Handler]("test", call, HandlerFunc(notFound), HandlerFunc(trace), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
		return &routertest.TestRouter[Handler]{
			ServeHTTP: router.ServeHTTP,
			Handle:    func(pattern string, h Handler, methods ...string) { router.Handle(pattern, h, nil, methods...) },
			Clean:     router.Clean,
			Remove:    router.Remove,
			URL:       router.URL,
		}
	})
	t.Bench(b, h)
}
