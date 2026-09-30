// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package routertest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"runtime/metrics"
	"testing"

	"github.com/issue9/mux/v10"
)

// Bench 执行所有的性能测试
//
// h 表示路由的处理函数，只要向终端输出 URL.Path 值即可，
// 以 T 的类型为 http.HandlerFunc 为例：
//
//	func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte(r.URL.Path))
//	}
func (t *Tester[T]) Bench(b *testing.B, h T) {
	allocs, r := t.load(h)
	fmt.Printf("\n加载 %d 条路由总共占用 %d KB\n", len(apis), allocs/1024)

	b.Run("URL", func(b *testing.B) {
		t.benchURL(b, r)
	})

	b.Run("Serve", func(b *testing.B) {
		t.benchServeHTTP(b, r)
	})

	b.Run("AddServe", func(b *testing.B) {
		t.benchAddAndServeHTTP(b, r, h)
	})
}

func (t *Tester[T]) benchURL(b *testing.B, router *TestRouter[T]) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		url, err := router.URL(true, api.pattern, api.ps)
		if err != nil {
			b.Error(err)
		}
		if url != api.test {
			b.Errorf("URL 出错，位于 %s", api.pattern)
		}
	}
}

func (t *Tester[T]) benchAddAndServeHTTP(b *testing.B, router *TestRouter[T], h T) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(api.method, api.test, nil)

		router.Remove(api.pattern, api.method)
		router.Handle(api.pattern, h, api.method)
		router.ServeHTTP(w, r)

		if w.Body.String() != r.URL.Path {
			b.Errorf("%s:%s", w.Body.String(), r.URL.Path)
		}
	}
}

func (t *Tester[T]) benchServeHTTP(b *testing.B, router *TestRouter[T]) {
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

// 加载路由
//
// 返回分配内存大小和测试路由对象
func (t *Tester[T]) load(h T) (uint64, *TestRouter[T]) {
	sample := make([]metrics.Sample, 1)
	sample[0].Name = "/gc/heap/allocs:bytes"

	runtime.GC()
	metrics.Read(sample)
	before := sample[0].Value.Uint64()

	r := t.loader(mux.WithLock(true))
	for _, api := range apis {
		r.Handle(api.pattern, h, api.method)
	}

	metrics.Read(sample)
	after := sample[0].Value.Uint64()

	return after - before, r
}
