// SPDX-License-Identifier: MIT

package mux_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/routertest"
)

func BenchmarkServeFile(b *testing.B) {
	fsys := os.DirFS("./")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/assets/mux.go", nil)
		err := mux.ServeFile(fsys, "mux.go", "", w, r)
		if err != nil {
			b.Errorf("测试出错，返回了以下错误 %s", err)
		}
	}
}

func BenchmarkDefaultRouter(b *testing.B) {
	h := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(r.URL.Path)); err != nil {
			panic(err)
		}
	}

	t := routertest.NewTester[http.Handler](mux.DefaultCall)
	t.Bench(b, http.HandlerFunc(h))
}

func BenchmarkContextRouter(b *testing.B) {
	h := func(c *ctx) {
		if _, err := c.W.Write([]byte(c.R.URL.Path)); err != nil {
			panic(err)
		}
	}

	t := routertest.NewTester[ctxHandlerFunc](contextCall)
	t.Bench(b, h)
}
