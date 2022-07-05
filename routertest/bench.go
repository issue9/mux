// SPDX-License-Identifier: MIT

package routertest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/metrics"
	"testing"

	"github.com/issue9/mux/v7"
)

// Bench 执行所有的性能测试
//
// h 表示路由的处理函数，只要向终端输出 URL.Path 值即可，
// 以 T 的类型为 http.HandlerFunc 为例：
//  func(w http.ResponseWriter, r *http.Request) {
//      w.Write([]byte(r.URL.Path))
//  }
func (t *Tester[T]) Bench(b *testing.B, h T) {
	allocs := t.calcMemStats(h)
	fmt.Printf("\n加载 %d 条路由总共占用 %d KB\n", len(apis), allocs/1024)

	b.Run("URL", func(b *testing.B) {
		t.benchURL(b, h)
	})

	b.Run("Add", func(b *testing.B) {
		t.benchAddAndServeHTTP(b, h)
	})

	b.Run("Serve", func(b *testing.B) {
		t.benchServeHTTP(b, h)
	})
}

func (t *Tester[T]) benchURL(b *testing.B, h T) {
	const domain = "https://github.com"

	router := mux.NewRouterOf("test", t.c, t.notFound, t.m, t.o, &mux.Options{
		Lock:      true,
		URLDomain: domain,
	})
	for _, api := range apis {
		router.Handle(api.pattern, h, api.method)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		url, err := router.URL(true, api.pattern, api.ps)
		if err != nil {
			b.Error(err)
		}
		if url != domain+api.test {
			b.Errorf("URL 出错，位于 %s", api.pattern)
		}
	}
}

func (t *Tester[T]) benchAddAndServeHTTP(b *testing.B, h T) {
	router := mux.NewRouterOf("test", t.c, t.notFound, t.m, t.o, &mux.Options{
		Lock: true,
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(api.method, api.test, nil)

		router.Handle(api.pattern, h, api.method)
		router.ServeHTTP(w, r)
		router.Remove(api.pattern, api.method)

		if w.Body.String() != r.URL.Path {
			b.Errorf("%s:%s", w.Body.String(), r.URL.Path)
		}
	}
}

func (t *Tester[T]) benchServeHTTP(b *testing.B, h T) {
	router := mux.NewRouterOf("test", t.c, t.notFound, t.m, t.o, nil)
	for _, api := range apis {
		router.Handle(api.pattern, h, api.method)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(api.method, api.test, nil)
		router.ServeHTTP(w, r)

		if w.Body.String() != r.URL.Path {
			b.Errorf("%s:%s", w.Body.String(), r.URL.Path)
		}
	}
}

func (t *Tester[T]) calcMemStats(h T) uint64 {
	return calcMemStats(func() {
		r := mux.NewRouterOf("test", t.c, t.notFound, t.m, t.o, &mux.Options{Lock: true})
		for _, api := range apis {
			r.Handle(api.pattern, h, api.method)
		}
	})
}

func calcMemStats(load func()) uint64 {
	sample := make([]metrics.Sample, 1)
	sample[0].Name = "/gc/heap/allocs:bytes"

	runtime.GC()
	metrics.Read(sample)
	before := sample[0].Value.Uint64()

	load()

	metrics.Read(sample)
	after := sample[0].Value.Uint64()

	return after - before
}
