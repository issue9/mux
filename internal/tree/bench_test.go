// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/types"
)

func BenchmarkTree_Handler(b *testing.B) {
	a := assert.New(b, false)
	tree := NewTestTree(a, true, nil, syntax.NewInterceptors())

	// 添加路由项
	a.NotError(tree.Add("/", rest.BuildHandler(a, 201, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, 202, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, 203, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", rest.BuildHandler(a, 204, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}", rest.BuildHandler(a, 205, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}/author", rest.BuildHandler(a, 206, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/page/{page:\\d*}", rest.BuildHandler(a, 207, "", nil), nil, http.MethodGet)) // 可选的正则节点
	a.NotError(tree.Add("/posts/{id}/{page}/author", rest.BuildHandler(a, 208, "", nil), nil, http.MethodGet))

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

	l := len(paths)

	// 确保数据是正确返回的
	for i := range l {
		ctx := types.NewContext()
		index := i % len(paths)
		ctx.Path = paths[index]
		node, h, ok := tree.Handler(ctx, http.MethodGet)
		a.True(ok).
			NotNil(node).
			NotNil(h)
	}

	b.ResetTimer()
	ctx := types.NewContext()
	for i := range b.N {
		ctx.Reset()
		index := i % l
		ctx.Path = paths[index]
		tree.Handler(ctx, http.MethodGet)
	}
}
