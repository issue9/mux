// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package std

import (
	"net/http"
	"testing"

	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/routertest"
)

func BenchmarkRouter(b *testing.B) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.URL.Path)); err != nil {
			panic(err)
		}
	}

	t := routertest.NewTester[http.Handler](func(o ...mux.Option) *routertest.TestRouter[http.Handler] {
		router := mux.NewRouter[http.Handler]("test", call, http.NotFoundHandler(), http.HandlerFunc(trace), methodNotAllowedBuilder, optionsHandlerBuilder, o...)
		return &routertest.TestRouter[http.Handler]{
			ServeHTTP: router.ServeHTTP,
			Handle:    func(pattern string, h http.Handler, methods ...string) { router.Handle(pattern, h, nil, methods...) },
			Clean:     router.Clean,
			Remove:    router.Remove,
			URL:       router.URL,
		}
	})

	t.Bench(b, http.HandlerFunc(h))
}
