// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
		tree: New(),
		a:    a,
	}
}

// 添加一条路由项。code 表示该路由项返回的报头，
// 测试路由项的 code 需要唯一，之后也是通过此值来判断其命中的路由项。
func (t *tester) add(method, pattern string, code int) {
	t.a.NotError(t.tree.Add(pattern, buildHandler(code), method))
}

// 验证按照指定的 method 和 path 访问，是否会返回相同的 code 值，
// 若是，则返回该节点以及对应的参数。
func (t *tester) handler(method, path string, code int) (http.Handler, params.Params) {
	hs, ps := t.tree.Route(path)
	t.a.NotNil(hs)

	h := hs.Handler(method)
	t.a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(method, path, nil)
	h.ServeHTTP(w, r)
	t.a.Equal(w.Code, code)

	return h, ps
}

// 验证指定的路径是否匹配正确的路由项，通过 code 来确定，并返回该节点的实例。
func (t *tester) matchTrue(method, path string, code int) {
	h, _ := t.handler(method, path, code)
	t.a.NotNil(h)
}

// 验证指定的路径返回的参数是否正确
func (t *tester) paramsTrue(method, path string, code int, params map[string]string) {
	_, ps := t.handler(method, path, code)
	t.a.Equal(ps, params)
}

// 测试 tree.methodIndex 是否正确
func (t *tester) optionsTrue(path, options string) {
	hs, ps := t.tree.Route(path)
	t.a.Nil(ps).NotNil(hs)

	h := hs.Handler(http.MethodOptions)
	t.a.NotNil(h)

	w := httptest.NewRecorder()
	u, err := url.Parse(path)
	t.a.NotError(err).NotNil(u)
	r, err := http.NewRequest(http.MethodOptions, path, nil) // httptest.NewRequest 不支持 path=''
	t.a.NotError(err).NotNil(r)
	h.ServeHTTP(w, r)
	t.a.Equal(w.Code, http.StatusOK)

	t.a.Equal(w.Header().Get("Allow"), options)
}

func TestTree_match(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/", 201)
	test.add(http.MethodGet, "/posts/{id}", 202)
	test.add(http.MethodGet, "/posts/{id}/author", 203)
	test.add(http.MethodGet, "/posts/1/author", 204)
	test.add(http.MethodGet, "/posts/{id:\\d+}", 205)
	test.add(http.MethodGet, "/posts/{id:\\d+}/author", 206)
	test.add(http.MethodGet, "/page/{page:\\d*}", 207) // 可选的正则节点
	test.add(http.MethodGet, "/posts/{id}/{page}/author", 208)

	test.matchTrue(http.MethodGet, "/", 201)
	test.matchTrue(http.MethodGet, "/posts/1", 205)             // 正则
	test.matchTrue(http.MethodGet, "/posts/2", 205)             // 正则
	test.matchTrue(http.MethodGet, "/posts/2/author", 206)      // 正则
	test.matchTrue(http.MethodGet, "/posts/1/author", 204)      // 字符串
	test.matchTrue(http.MethodGet, "/posts/1.html", 202)        // 命名参数
	test.matchTrue(http.MethodGet, "/posts/1.html/page", 202)   // 命名参数
	test.matchTrue(http.MethodGet, "/posts/2.html/author", 203) // 命名参数
	test.matchTrue(http.MethodGet, "/page/", 207)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 208) // 若 {id} 可匹配任意字符，此条也可匹配 3

	// 测试 digit 和 \\d 是否正常
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 201)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202)

	// 测试 digit 和 named 是否正常
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id}/author", 201)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202)

	// 测试 digit 和 \\d 和 named 三者顺序是否正常
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id}/author", 201)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 202)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 203)
	test.matchTrue(http.MethodGet, "/posts/1/author", 203)

	// 以斜框结尾，是否能正常访问
	test = newTester(a)
	test.add(http.MethodGet, "/posts/{id}/", 201)
	test.add(http.MethodGet, "/posts/{id}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/", 201)
	test.matchTrue(http.MethodGet, "/posts/1.html/", 201)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202)

	// 以 - 作为路径分隔符
	test = newTester(a)
	test.add(http.MethodGet, "/posts-{id}", 201)
	test.add(http.MethodGet, "/posts-{id}-author", 202)
	test.matchTrue(http.MethodGet, "/posts-1.html", 201)
	test.matchTrue(http.MethodGet, "/posts-1-author", 202)

	test = newTester(a)
	test.add(http.MethodGet, "/admin/{path}", 201)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}", 202)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile", 203)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile/{type:\\d+}", 204)
	test.matchTrue(http.MethodGet, "/admin/index.html", 201)
	test.matchTrue(http.MethodGet, "/admin/items/1", 202)
	test.matchTrue(http.MethodGet, "/admin/items/1/profile", 203)
	test.matchTrue(http.MethodGet, "/admin/items/1/profile/1", 204)

	// 测试 indexes 功能
	test = newTester(a)
	test.add(http.MethodGet, "/admin/1", 201)
	test.add(http.MethodGet, "/admin/2", 202)
	test.add(http.MethodGet, "/admin/3", 203)
	test.add(http.MethodGet, "/admin/4", 204)
	test.add(http.MethodGet, "/admin/5", 205)
	test.add(http.MethodGet, "/admin/6", 206)
	test.add(http.MethodGet, "/admin/7", 207)
	test.add(http.MethodGet, "/admin/{id}", 220)
	c0 := test.tree.node.children[0]
	a.Equal(len(c0.indexes), 7)
	a.Equal(c0.segment.Value, "/admin/")
	// /admin/5index.html 即部分匹配 /admin/5，更匹配 /admin/{id}，测试是否正确匹配
	test.matchTrue(http.MethodGet, "/admin/5index.html", 220)

	// 测试非英文字符 indexes 功能
	test = newTester(a)
	test.add(http.MethodGet, "/中文/1", 201)
	test.add(http.MethodGet, "/中文/2", 202)
	test.add(http.MethodGet, "/中文/3", 203)
	test.add(http.MethodGet, "/中文/4", 204)
	test.add(http.MethodGet, "/中文/5", 205)
	test.add(http.MethodGet, "/中文/6", 206)
	test.add(http.MethodGet, "/中文/7", 207)
	test.add(http.MethodGet, "/中文/{id}", 220)
	c0 = test.tree.node.children[0]
	a.Equal(len(c0.indexes), 7)
	a.Equal(c0.segment.Value, "/中文/")
	// /中文/5index.html 即部分匹配 /中文/5，更匹配 /中文/{id}，测试是否正确匹配
	test.matchTrue(http.MethodGet, "/中文/5index.html", 220)

	// 测试非英文字符 indexes 功能，汉字中的相同部分会被提取到上一级。
	test = newTester(a)
	test.add(http.MethodGet, "/中文/1", 201)
	test.add(http.MethodGet, "/中文/2", 202)
	test.add(http.MethodGet, "/中文/3", 203)
	test.add(http.MethodGet, "/丽文/4", 204)
	test.add(http.MethodGet, "/丽文/5", 205)
	test.add(http.MethodGet, "/丽文/6", 206)
	test.add(http.MethodGet, "/丽文/7", 207)
	test.add(http.MethodGet, "/中文/{id}", 220)
	c0 = test.tree.node.children[0]
	a.Equal(len(c0.indexes), 0)
	a.Equal(c0.segment.Value, "/\xe4\xb8")
	// /中文/5index.html 即部分匹配 /中文/5，更匹配 /中文/{id}，测试是否正确匹配
	test.matchTrue(http.MethodGet, "/中文/5index.html", 220)

	test = newTester(a)
	test.add(http.MethodGet, "/admin/1", 201)
	test.matchTrue(http.MethodGet, "/admin/1", 201)
	test.matchTrue(http.MethodHead, "/admin/1", 201) // 同 GET 的状态码
	test.optionsTrue("", "GET, HEAD, OPTIONS")
	test.optionsTrue("*", "GET, HEAD, OPTIONS")

	// 动态添加其它方法，会改变 OPTIONS 值
	test.add(http.MethodPost, "/admin/1", 201)
	test.optionsTrue("", "GET, HEAD, OPTIONS, POST")
}

func TestTree_Params(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 201)                       // 命名
	test.add(http.MethodGet, "/posts/{id}/author/{action}/", 202)      // 命名
	test.add(http.MethodGet, "/posts/{id:\\d+}", 203)                  // 正则
	test.add(http.MethodGet, "/posts/{id:\\d+}/author/{action}/", 204) // 正则

	// 正则
	test.paramsTrue(http.MethodGet, "/posts/1", 203, map[string]string{"id": "1"})
	test.paramsTrue(http.MethodGet, "/posts/1/author/profile/", 204, map[string]string{"id": "1", "action": "profile"})

	// 命名
	test.paramsTrue(http.MethodGet, "/posts/1.html", 201, map[string]string{"id": "1.html"})
	test.paramsTrue(http.MethodGet, "/posts/1.html/author/profile/", 202, map[string]string{"id": "1.html", "action": "profile"})
}

func TestTreeCN(t *testing.T) {
	a := assert.New(t)
	test := newTester(a)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 201) // 命名
	test.add(http.MethodGet, "/文章/{编号}", 202)    // 中文

	test.matchTrue(http.MethodGet, "/posts/1", 201)
	test.matchTrue(http.MethodGet, "/文章/1", 202)
	test.paramsTrue(http.MethodGet, "/文章/1.html", 202, map[string]string{"编号": "1.html"})
}

func TestTree_Clean(t *testing.T) {
	a := assert.New(t)
	tree := New()

	addNode := func(p string, code int, methods ...string) {
		a.NotError(tree.Add(p, buildHandler(code), methods...))
	}

	addNode("/", 201, http.MethodGet)
	addNode("/posts/1/author", 201, http.MethodGet)
	addNode("/posts/{id}", 201, http.MethodGet)
	addNode("/posts/{id}/author", 201, http.MethodGet)
	addNode("/posts/{id}/{author:\\w+}/profile", 201, http.MethodGet)

	a.Equal(tree.node.len(), 6) // 包含了 OPTIONS *

	tree.Clean("/posts/{id")
	a.Equal(tree.node.len(), 3)

	tree.Clean("")
	a.Equal(tree.node.len(), 1)
}

func TestTree_Add_Remove(t *testing.T) {
	a := assert.New(t)

	tree := New()
	a.NotNil(tree)

	a.NotError(tree.Add("/", buildHandler(http.StatusAccepted), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", buildHandler(http.StatusAccepted), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(http.StatusAccepted), http.MethodGet, http.MethodPut, http.MethodPost))
	a.NotError(tree.Add("/posts/1/author", buildHandler(http.StatusAccepted), http.MethodGet))
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

	tree = New()
	a.NotNil(tree)
	a.NotError(tree.Add("/path", buildHandler(201)))
	node := tree.node.find("/path")
	a.Equal(len(Methods), len(node.handlers))
	a.Equal(len(Methods), len(node.Methods()))
	a.Equal(node.Options(), strings.Join(node.Methods(), ", "))

	// head

	tree = New()
	a.NotNil(tree)
	a.NotError(tree.Add("/path", buildHandler(http.StatusAccepted), http.MethodGet))
	node = tree.node.find("/path")
	a.Equal(3, len(node.handlers)).
		NotNil(node.handlers[http.MethodHead]).
		NotNil(node.handlers[http.MethodOptions])

	tree.Remove("/path", http.MethodGet)
	a.Equal(0, len(node.handlers))

	// error

	tree = New()
	a.NotNil(tree)
	a.ErrorString(tree.Add("/path", buildHandler(http.StatusAccepted), http.MethodHead), "无法手动添加 OPTIONS/HEAD 请求方法")
	a.ErrorString(tree.Add("/path", buildHandler(http.StatusAccepted), "NOT-SUPPORTED"), "NOT-SUPPORTED")
	a.NotError(tree.Add("/path", buildHandler(http.StatusAccepted), http.MethodDelete))
	a.ErrorString(tree.Add("/path", buildHandler(http.StatusAccepted), http.MethodDelete), http.MethodDelete)
	tree.Remove("/path", http.MethodOptions) // remove options 不发生任何操作
	a.Equal(tree.node.find("/path").Options(), "DELETE, OPTIONS")

	a.ErrorString(tree.Add("/path/{id}/path/{id:\\d+}", buildHandler(1), http.MethodHead), "存在相同名称的路由参数")
	a.ErrorString(tree.Add("/path/{id}{id2:\\d+}", buildHandler(1), http.MethodHead), "两个命名参数不能连续出现")
}

func TestTree_Routes(t *testing.T) {
	a := assert.New(t)
	tree := New()
	a.NotNil(tree)

	a.NotError(tree.Add("/", buildHandler(http.StatusOK), http.MethodGet))
	a.NotError(tree.Add("/posts", buildHandler(http.StatusOK), http.MethodGet, http.MethodPost))
	a.NotError(tree.Add("/posts/{id}", buildHandler(http.StatusOK), http.MethodGet, http.MethodPut))
	a.NotError(tree.Add("/posts/{id}/author", buildHandler(http.StatusOK), http.MethodGet))

	routes := tree.Routes()
	a.Equal(routes, map[string][]string{
		"*":                  {http.MethodOptions},
		"/":                  {http.MethodGet, http.MethodHead, http.MethodOptions},
		"/posts":             {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost},
		"/posts/{id}":        {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut},
		"/posts/{id}/author": {http.MethodGet, http.MethodHead, http.MethodOptions},
	})
}
