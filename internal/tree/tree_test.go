// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/method"
)

func buildHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func TestTree_Add_Remove(t *testing.T) {
	a := assert.New(t)
	tree := New()

	a.NotError(tree.Add("/", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(1), http.MethodGet, http.MethodPut, http.MethodPost))
	a.NotError(tree.Add("/posts/1/author", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", buildHandler(1), http.MethodGet))
	tree.print(0)

	a.NotEmpty(tree.find("/posts/1/author").handlers.handlers)
	a.NotError(tree.Remove("/posts/1/author", http.MethodGet))
	a.Nil(tree.find("/posts/1/author"))

	a.NotError(tree.Remove("/posts/{id}/author", http.MethodGet)) // 只删除 GET
	a.NotNil(tree.find("/posts/{id}/author"))
	a.NotError(tree.Remove("/posts/{id}/author", method.Supported...)) // 删除所有请求方法
	a.Nil(tree.find("/posts/{id}/author"))
	a.Error(tree.Remove("/posts/{id}/author", method.Supported...)) // 删除已经不存在的节点

}
