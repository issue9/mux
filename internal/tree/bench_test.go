// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/mux/v7/internal/syntax"
	"github.com/issue9/mux/v7/types"
)

func BenchmarkTree_Match(b *testing.B) {
	a := assert.New(b, false)
	tree := NewTestTree(a, true, syntax.NewInterceptors())

	// 添加路由项
	a.NotError(tree.Add("/", rest.BuildHandler(a, 201, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, 202, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, 203, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", rest.BuildHandler(a, 204, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}", rest.BuildHandler(a, 205, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}/author", rest.BuildHandler(a, 206, "", nil), http.MethodGet))
	a.NotError(tree.Add("/page/{page:\\d*}", rest.BuildHandler(a, 207, "", nil), http.MethodGet)) // 可选的正则节点
	a.NotError(tree.Add("/posts/{id}/{page}/author", rest.BuildHandler(a, 208, "", nil), http.MethodGet))

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
		p := types.NewContext()
		p.Path = paths[index]
		h := tree.match(p)
		a.True(len(h.handlers) > 0)
	}
}

func BenchmarkTree_ServeHTTP(b *testing.B) {
	a := assert.New(b, false)
	tree := NewTestTree(a, true, syntax.NewInterceptors())

	// 添加路由项
	a.NotError(tree.Add("/", rest.BuildHandler(a, 201, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, 202, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, 203, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", rest.BuildHandler(a, 204, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}", rest.BuildHandler(a, 205, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}/author", rest.BuildHandler(a, 206, "", nil), http.MethodGet))
	a.NotError(tree.Add("/page/{page:\\d*}", rest.BuildHandler(a, 207, "", nil), http.MethodGet)) // 可选的正则节点
	a.NotError(tree.Add("/posts/{id}/{page}/author", rest.BuildHandler(a, 208, "", nil), http.MethodGet))

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
		p := types.NewContext()
		p.Path = paths[index]
		h := tree.match(p)
		hh := h.handlers[http.MethodGet]

		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, paths[index], nil)
		a.NotError(err).NotNil(r)
		hh.ServeHTTP(w, r)
		a.Equal(w.Result().StatusCode, index+201)
	}
}
