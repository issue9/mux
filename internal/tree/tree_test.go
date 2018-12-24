// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/handlers"
	"github.com/issue9/mux/params"
)

func buildHandler(code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

type tester struct {
	tree *Tree
	a    *assert.Assertion
}

func newTester(a *assert.Assertion) *tester {
	return &tester{
		tree: New(false, false),
		a:    a,
	}
}

// 添加一条路由项。code 表示该路由项返回的报头，
// 测试路由项的 code 需要唯一，之后也是通过此值来判断其命中的路由项。
func (n *tester) add(method, pattern string, code int) {
	nn, err := n.tree.getNode(pattern)
	n.a.NotError(err).NotNil(nn)

	if nn.handlers == nil {
		nn.handlers = handlers.New(false, false)
	}

	nn.handlers.Add(buildHandler(code), method)
}

// 验证按照指定的 method 和 path 访问，是否会返回相同的 code 值，
// 若是，则返回该节点以及对应的参数。
func (n *tester) handler(method, path string, code int) (http.Handler, params.Params) {
	hs, ps := n.tree.Handler(path)
	n.a.NotNil(ps).NotNil(hs)

	h := hs.Handler(method)
	n.a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	h.ServeHTTP(w, r)
	n.a.Equal(w.Code, code)

	return h, ps
}

// 验证指定的路径是否匹配正确的路由项，通过 code 来确定，并返回该节点的实例。
func (n *tester) matchTrue(method, path string, code int) {
	h, _ := n.handler(method, path, code)
	n.a.NotNil(h)
}

// 验证指定的路径返回的参数是否正确
func (n *tester) paramsTrue(method, path string, code int, params map[string]string) {
	_, ps := n.handler(method, path, code)
	n.a.Equal(ps, params)
}

// 验证 node.URL 的正确性
func (n *tester) urlTrue(pattern string, params map[string]string, url string) {
	u, err := n.tree.URL(pattern, params)
	n.a.NotError(err)
	n.a.Equal(u, url)
}

func TestTree_match(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/", 1)
	test.add(http.MethodGet, "/posts/{id}", 2)
	test.add(http.MethodGet, "/posts/{id}/author", 3)
	test.add(http.MethodGet, "/posts/1/author", 4)
	test.add(http.MethodGet, "/posts/{id:\\d+}", 5)
	test.add(http.MethodGet, "/posts/{id:\\d+}/author", 6)
	test.add(http.MethodGet, "/page/{page:\\d*}", 7) // 可选的正则节点
	test.add(http.MethodGet, "/posts/{id}/{page}/author", 8)
	//test.n.print(0)

	test.matchTrue(http.MethodGet, "/", 1)
	test.matchTrue(http.MethodGet, "/posts/1", 5)             // 正则
	test.matchTrue(http.MethodGet, "/posts/2", 5)             // 正则
	test.matchTrue(http.MethodGet, "/posts/2/author", 6)      // 正则
	test.matchTrue(http.MethodGet, "/posts/1/author", 4)      // 字符串
	test.matchTrue(http.MethodGet, "/posts/1.html", 2)        // 命名参数
	test.matchTrue(http.MethodGet, "/posts/1.html/page", 2)   // 命名参数
	test.matchTrue(http.MethodGet, "/posts/2.html/author", 3) // 命名参数
	test.matchTrue(http.MethodGet, "/page/", 7)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 8) // 若 {id} 可匹配任意字符，此条也可匹配 3

	// 以斜框结尾，是否能正常访问
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id}/", 1)
	test.add(http.MethodGet, "/posts/{id}/author", 2)
	test.matchTrue(http.MethodGet, "/posts/1/", 1)
	test.matchTrue(http.MethodGet, "/posts/1.html/", 1)
	test.matchTrue(http.MethodGet, "/posts/1/author", 2)

	// 以 - 作为路径分隔符
	test = newTester(a)
	test.add(http.MethodGet, "/posts-{id}", 1)
	test.add(http.MethodGet, "/posts-{id}-author", 2)
	test.matchTrue(http.MethodGet, "/posts-1.html", 1)
	test.matchTrue(http.MethodGet, "/posts-1-author", 2)

	test = newTester(a)
	test.add(http.MethodGet, "/admin/{path}", 1)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}", 2)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile", 3)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile/{type:\\d+}", 4)
	test.matchTrue(http.MethodGet, "/admin/index.html", 1)
	test.matchTrue(http.MethodGet, "/admin/items/1", 2)
	test.matchTrue(http.MethodGet, "/admin/items/1/profile", 3)
	test.matchTrue(http.MethodGet, "/admin/items/1/profile/1", 4)
	test.tree.Trace(os.Stdout, "/admin/items/1/profile/1")
}

func TestTree_Params(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 1)                       // 命名
	test.add(http.MethodGet, "/posts/{id}/author/{action}/", 2)      // 命名
	test.add(http.MethodGet, "/posts/{id:\\d+}", 3)                  // 正则
	test.add(http.MethodGet, "/posts/{id:\\d+}/author/{action}/", 4) // 正则

	// 正则
	test.paramsTrue(http.MethodGet, "/posts/1", 3, map[string]string{"id": "1"})
	test.paramsTrue(http.MethodGet, "/posts/1/author/profile/", 4, map[string]string{"id": "1", "action": "profile"})

	// 命名
	test.paramsTrue(http.MethodGet, "/posts/1.html", 1, map[string]string{"id": "1.html"})
	test.paramsTrue(http.MethodGet, "/posts/1.html/author/profile/", 2, map[string]string{"id": "1.html", "action": "profile"})
}

func TestTree_URL(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 1)                       // 命名
	test.add(http.MethodGet, "/posts/{id}/author/{action}/", 2)      // 命名
	test.add(http.MethodGet, "/posts/{id:\\d+}", 3)                  // 正则
	test.add(http.MethodGet, "/posts/{id:\\d+}/author/{action}/", 4) // 正则

	test.urlTrue("/posts/{id:\\d+}", map[string]string{"id": "100"}, "/posts/100")
	test.urlTrue("/posts/{id:\\d+}/author/{action}/", map[string]string{"id": "100", "action": "p"}, "/posts/100/author/p/")
	test.urlTrue("/posts/{id}", map[string]string{"id": "100.htm"}, "/posts/100.htm")
	test.urlTrue("/posts/{id}/author/{action}/", map[string]string{"id": "100.htm", "action": "p"}, "/posts/100.htm/author/p/")
}

func TestTreeCN(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 1) // 命名
	test.add(http.MethodGet, "/文章/{编号}", 2)    // 中文

	test.matchTrue(http.MethodGet, "/posts/1", 1)
	test.matchTrue(http.MethodGet, "/文章/1", 2)
	test.paramsTrue(http.MethodGet, "/文章/1.html", 2, map[string]string{"编号": "1.html"})
	test.urlTrue("/文章/{编号}", map[string]string{"编号": "100.htm"}, "/文章/100.htm")
}

func TestTree_Clean(t *testing.T) {
	a := assert.New(t)
	tree := New(false, false)

	addNode := func(p string, code int, methods ...string) {
		nn, err := tree.getNode(p)
		a.NotError(err).NotNil(nn)

		if nn.handlers == nil {
			nn.handlers = handlers.New(false, false)
		}

		a.NotError(nn.handlers.Add(buildHandler(code), methods...))
	}

	addNode("/", 1, http.MethodGet)
	addNode("/posts/1/author", 1, http.MethodGet)
	addNode("/posts/{id}", 1, http.MethodGet)
	addNode("/posts/{id}/author", 1, http.MethodGet)
	addNode("/posts/{id}/{author:\\w+}/profile", 1, http.MethodGet)

	a.Equal(tree.len(), 5)

	tree.Clean("/posts/{id")
	a.Equal(tree.len(), 2)

	tree.Clean("")
	a.Equal(tree.len(), 0)
}

func TestTree_Add_Remove(t *testing.T) {
	a := assert.New(t)
	tree := New(false, false)

	a.NotError(tree.Add("/", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(1), http.MethodGet, http.MethodPut, http.MethodPost))
	a.NotError(tree.Add("/posts/1/author", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", buildHandler(1), http.MethodGet))

	a.True(tree.find("/posts/1/author").handlers.Len() > 0)
	a.NotError(tree.Remove("/posts/1/author", http.MethodGet))
	a.Nil(tree.find("/posts/1/author"))

	a.NotError(tree.Remove("/posts/{id}/author", http.MethodGet)) // 只删除 GET
	a.NotNil(tree.find("/posts/{id}/author"))
	a.NotError(tree.Remove("/posts/{id}/author")) // 删除所有请求方法
	a.Nil(tree.find("/posts/{id}/author"))
	a.Error(tree.Remove("/posts/{id}/author")) // 删除已经不存在的节点
}

func TestTree_SetAllow(t *testing.T) {
	a := assert.New(t)
	tree := New(false, false)

	a.NotError(tree.Add("/options", buildHandler(1), http.MethodGet, http.MethodDelete, http.MethodPatch))
	tree.SetAllow("/options", "TEST1,TEST2")
	n, err := tree.getNode("/options")
	a.NotError(err).NotNil(n)
	a.Equal(n.handlers.Options(), "TEST1,TEST2")

	tree.Remove("/options", http.MethodGet)
	a.Equal(n.handlers.Options(), "TEST1,TEST2")

	// 删除完之后，清空 allow
	tree.Remove("/options")
	a.Equal(n.handlers.Options(), "")
}
