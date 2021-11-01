// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkTree_Handler(b *testing.B) {
	a := assert.New(b)
	tree := New()

	// 添加路由项
	a.NotError(tree.Add("/", buildHandler(http.StatusAccepted), http.MethodGet))
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
