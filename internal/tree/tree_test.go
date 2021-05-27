// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/params"
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
		tree: New(false),
		a:    a,
	}
}

// 添加一条路由项。code 表示该路由项返回的报头，
// 测试路由项的 code 需要唯一，之后也是通过此值来判断其命中的路由项。
func (n *tester) add(method, pattern string, code int) {
	n.a.NotError(n.tree.Add(pattern, buildHandler(code), method))
}

// 验证按照指定的 method 和 path 访问，是否会返回相同的 code 值，
// 若是，则返回该节点以及对应的参数。
func (n *tester) handler(method, path string, code int) (http.Handler, params.Params) {
	hs, ps := n.tree.Route(path)
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

	// 测试 digit 和 \\d 是否正常
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 1)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 2)
	test.matchTrue(http.MethodGet, "/posts/1/author", 2)

	// 测试 digit 和 named 是否正常
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id}/author", 1)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 2)
	test.matchTrue(http.MethodGet, "/posts/1/author", 2)

	// 测试 digit 和 \\d 和 named 三者顺序是否正常
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id}/author", 1)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 2)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 3)
	test.matchTrue(http.MethodGet, "/posts/1/author", 3)

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

	// 测试 indexes 功能
	test = newTester(a)
	test.add(http.MethodGet, "/admin/1", 1)
	test.add(http.MethodGet, "/admin/2", 2)
	test.add(http.MethodGet, "/admin/3", 3)
	test.add(http.MethodGet, "/admin/4", 4)
	test.add(http.MethodGet, "/admin/5", 5)
	test.add(http.MethodGet, "/admin/6", 6)
	test.add(http.MethodGet, "/admin/7", 7)
	test.add(http.MethodGet, "/admin/{id}", 20)
	a.Equal(len(test.tree.node.children[0].indexes), 7)
	// /admin/5index.html 即部分匹配 /admin/5，更匹配 /admin/{id}，测试是否正确匹配
	test.matchTrue(http.MethodGet, "/admin/5index.html", 20)
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

func TestTreeCN(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 1) // 命名
	test.add(http.MethodGet, "/文章/{编号}", 2)    // 中文

	test.matchTrue(http.MethodGet, "/posts/1", 1)
	test.matchTrue(http.MethodGet, "/文章/1", 2)
	test.paramsTrue(http.MethodGet, "/文章/1.html", 2, map[string]string{"编号": "1.html"})
}

func TestTree_Clean(t *testing.T) {
	a := assert.New(t)
	tree := New(false)

	addNode := func(p string, code int, methods ...string) {
		a.NotError(tree.Add(p, buildHandler(code), methods...))
	}

	addNode("/", 1, http.MethodGet)
	addNode("/posts/1/author", 1, http.MethodGet)
	addNode("/posts/{id}", 1, http.MethodGet)
	addNode("/posts/{id}/author", 1, http.MethodGet)
	addNode("/posts/{id}/{author:\\w+}/profile", 1, http.MethodGet)

	a.Equal(tree.node.len(), 5)

	tree.Clean("/posts/{id")
	a.Equal(tree.node.len(), 2)

	tree.Clean("")
	a.Equal(tree.node.len(), 0)
}

func TestTree_Add_Remove(t *testing.T) {
	a := assert.New(t)

	tree := New(false)
	a.NotNil(tree)

	a.NotError(tree.Add("/", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(1), http.MethodGet, http.MethodPut, http.MethodPost))
	a.NotError(tree.Add("/posts/1/author", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", buildHandler(1), http.MethodGet))

	a.NotEmpty(tree.node.find("/posts/1/author").handlers)
	tree.Remove("/posts/1/author", http.MethodGet)
	a.Nil(tree.node.find("/posts/1/author"))

	tree.Remove("/posts/{id}/author", http.MethodGet) // 只删除 GET
	a.NotNil(tree.node.find("/posts/{id}/author"))
	tree.Remove("/posts/{id}/author") // 删除所有请求方法
	a.Nil(tree.node.find("/posts/{id}/author"))
	tree.Remove("/posts/{id}/author") // 删除已经不存在的节点，不会报错，不发生任何事情

	// addAny

	tree = New(false)
	a.NotNil(tree)
	a.NotError(tree.Add("/path", buildHandler(1)))
	node := tree.node.find("/path")
	a.Equal(len(Methods), len(node.handlers))
	a.Equal(len(Methods), len(node.Methods()))
	a.Equal(node.Options(), strings.Join(node.Methods(), ", "))

	// head

	tree = New(false)
	a.NotNil(tree)
	a.NotError(tree.Add("/path", buildHandler(1), http.MethodGet))
	node = tree.node.find("/path")
	a.Equal(3, len(node.handlers)).
		NotNil(node.handlers[http.MethodHead]).
		NotNil(node.handlers[http.MethodOptions])

	tree.Remove("/path", http.MethodGet)
	a.Equal(0, len(node.handlers))

	tree = New(true)
	a.NotNil(tree)
	a.NotError(tree.Add("/path", buildHandler(1), http.MethodGet))
	node = tree.node.find("/path")
	a.Equal(2, len(node.handlers)).
		Nil(node.handlers[http.MethodHead]).
		NotNil(node.handlers[http.MethodOptions])

	tree.Remove("/path", http.MethodGet)
	a.Equal(0, len(node.handlers))

	// error

	tree = New(false)
	a.NotNil(tree)
	a.ErrorString(tree.Add("/path", buildHandler(1), http.MethodHead), "无法手动添加 OPTIONS/HEAD 请求方法")
	a.ErrorString(tree.Add("/path", buildHandler(1), "NOT-SUPPORTED"), "NOT-SUPPORTED")
	a.NotError(tree.Add("/path", buildHandler(1), http.MethodDelete))
	a.ErrorString(tree.Add("/path", buildHandler(1), http.MethodDelete), http.MethodDelete)
	tree.Remove("/path", http.MethodOptions) // remove options 不发生任何操作
	a.Equal(tree.node.find("/path").Options(), "DELETE, OPTIONS")

	a.ErrorString(tree.Add("/path/{id}/path/{id:\\d+}", buildHandler(1), http.MethodHead), "存在相同名称的路由参数")
	a.ErrorString(tree.Add("/path/{id}{id2:\\d+}", buildHandler(1), http.MethodHead), "两个命名参数不能连续出现")
}

func TestTree_Routes(t *testing.T) {
	a := assert.New(t)
	tree := New(false)
	a.NotNil(tree)

	a.NotError(tree.Add("/", buildHandler(http.StatusOK), http.MethodGet))
	a.NotError(tree.Add("/posts", buildHandler(http.StatusOK), http.MethodGet, http.MethodPost))
	a.NotError(tree.Add("/posts/{id}", buildHandler(http.StatusOK), http.MethodGet, http.MethodPut))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(http.StatusOK), http.MethodGet))

	routes := tree.Routes()
	a.Equal(routes, map[string][]string{
		"/":                  {http.MethodGet, http.MethodHead, http.MethodOptions},
		"/posts":             {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost},
		"/posts/{id}":        {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut},
		"/posts/{id}/author": {http.MethodGet, http.MethodHead, http.MethodOptions},
	})
}

func BenchmarkTree_Handler(b *testing.B) {
	a := assert.New(b)
	tree := New(false)

	// 添加路由项
	a.NotError(tree.Add("/", buildHandler(1), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(2), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(3), http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", buildHandler(4), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}", buildHandler(5), http.MethodGet))
	a.NotError(tree.Add("/posts/{id:\\d+}/author", buildHandler(6), http.MethodGet))
	a.NotError(tree.Add("/page/{page:\\d*}", buildHandler(7), http.MethodGet)) // 可选的正则节点
	a.NotError(tree.Add("/posts/{id}/{page}/author", buildHandler(8), http.MethodGet))

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
