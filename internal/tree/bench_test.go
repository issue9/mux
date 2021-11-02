// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkTree_Route(b *testing.B) {
	a := assert.New(b)
	tree := New()

	// 添加路由项
	a.NotError(tree.Add("/", buildHandler(201), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(202), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(203), http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", buildHandler(204), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}", buildHandler(205), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}/author", buildHandler(206), http.MethodGet))
	a.NotError(tree.Add("/page/{page:\\d*}", buildHandler(207), http.MethodGet)) // 可选的正则节点
	a.NotError(tree.Add("/posts/{id}/{page}/author", buildHandler(208), http.MethodGet))

	paths := map[int]string{
		0: "/",
		1: "/",
		2: "/posts/1.html/page",
		3: "/posts/2.html/author",
		4: "/posts/1/author",
		5: "/posts/2",
		6: "/posts/2/author",
		7: "/page/",
		8: "/posts/2.html/2/author",
	}

	for i := 0; i < b.N; i++ {
		index := i % len(paths)
		h, _ := tree.Route(paths[index])
		a.True(len(h.handlers) > 0)
	}
}

func BenchmarkTree_ServeHTTP(b *testing.B) {
	a := assert.New(b)
	tree := New()

	// 添加路由项
	a.NotError(tree.Add("/", buildHandler(201), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(202), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(203), http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", buildHandler(204), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}", buildHandler(205), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}/author", buildHandler(206), http.MethodGet))
	a.NotError(tree.Add("/page/{page:\\d*}", buildHandler(207), http.MethodGet)) // 可选的正则节点
	a.NotError(tree.Add("/posts/{id}/{page}/author", buildHandler(208), http.MethodGet))

	// 与上面路路依次相对，其 键名+201 即为其返回的状态码。
	paths := map[int]string{
		0: "/",
		1: "/posts/1.html/page",
		2: "/posts/2.html/author",
		3: "/posts/1/author",
		4: "/posts/2",
		5: "/posts/2/author",
		6: "/page/",
		7: "/posts/2.html/2/author",
	}

	for i := 0; i < b.N; i++ {
		index := i % len(paths)
		h, _ := tree.Route(paths[index])
		hh := h.handlers[http.MethodGet]

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, paths[index], nil)
		hh.ServeHTTP(w, r)
		a.Equal(w.Result().StatusCode, index+201)
	}
}
