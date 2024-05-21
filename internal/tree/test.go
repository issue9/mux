// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/issue9/assert/v4"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/internal/trace"
	"github.com/issue9/mux/v9/types"
)

func TestTrace(w http.ResponseWriter, r *http.Request) { trace.Trace(w, r, true) }

// NewTestTree 返回以 http.Handler 作为参数实例化的 Tree
func NewTestTree(a *assert.Assertion, lock bool, trace http.Handler, i *syntax.Interceptors) *Tree[http.Handler] {
	t := New("", lock, i, http.NotFoundHandler(), trace, BuildTestNodeHandlerFunc(http.StatusMethodNotAllowed), BuildTestNodeHandlerFunc(http.StatusOK))
	a.NotNil(t)
	return t
}

func BuildTestMiddleware(a *assert.Assertion, text string) types.Middleware[http.Handler] {
	return types.MiddlewareFunc[http.Handler](func(next http.Handler, _, _, _ string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r) // 先输出被包含的内容
			_, err := w.Write([]byte(text))
			a.NotError(err)
		})
	})
}

func BuildTestNodeHandlerFunc(status int) types.BuildNodeHandler[http.Handler] {
	return func(n types.Node) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(header.Allow, n.AllowHeader())
			w.WriteHeader(status)
		})
	}
}

// Print 向 w 输出树状结构
func (tree *Tree[T]) Print(w io.Writer) { tree.node.print(w, 0) }

func (n *node[T]) print(w io.Writer, deep int) {
	fmt.Fprintln(w, strings.Repeat(" ", deep*4), n.segment.Value)
	for _, child := range n.children {
		child.print(w, deep+1)
	}
}
