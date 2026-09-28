// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package routertest

import (
	"net/http"
	"testing"

	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/types"
)

// 测试 [mux.Router] 的主要性能
func BenchmarkRouter(b *testing.B) {
	trace := func(w http.ResponseWriter, r *http.Request) { mux.Trace(w, r, true) }
	m := func(node types.Node) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusMethodNotAllowed) })
	}
	o := func(node types.Node) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	}
	call := func(w http.ResponseWriter, r1 *http.Request, _ *types.Route, hf http.Handler) {
		hf.ServeHTTP(w, r1)
	}

	t := NewTester[http.Handler](call, http.NotFoundHandler(), http.HandlerFunc(trace), m, o)

	t.Bench(b, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(r.URL.Path)) }))
}
