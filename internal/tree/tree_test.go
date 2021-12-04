// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v5/internal/syntax"
)

type tester struct {
	tree *Tree
	a    *assert.Assertion
}

func newTester(a *assert.Assertion, lock bool) *tester {
	i := syntax.NewInterceptors()
	a.NotNil(i)

	i.Add(syntax.MatchDigit, "digit")

	return &tester{
		tree: New(lock, i),
		a:    a,
	}
}

// 添加一条路由项。code 表示该路由项返回的报头，
// 测试路由项的 code 需要唯一，之后也是通过此值来判断其命中的路由项。
func (t *tester) add(method, pattern string, code int) {
	t.a.TB().Helper()
	t.a.NotError(t.tree.Add(pattern, rest.BuildHandler(t.a, code, "", nil), method))
}

func (t *tester) addAmbiguous(pattern string) {
	t.a.TB().Helper()
	b := rest.BuildHandler(t.a, http.StatusOK, "", nil)
	t.a.ErrorString(t.tree.Add(pattern, b, http.MethodGet), "存在有歧义的节点")
}

// 验证按照指定的 method 和 path 访问，是否会返回相同的 code 值，
// 若是，则返回该节点以及对应的参数。
func (t *tester) handler(method, path string, code int) (http.Handler, *syntax.Params) {
	t.a.TB().Helper()

	hs, ps := t.tree.Route(path)
	t.a.NotNil(hs)

	h := hs.Handler(method)
	t.a.NotNil(h)

	w := httptest.NewRecorder()
	r, err := http.NewRequest(method, path, nil)
	t.a.NotError(err).NotNil(r)
	h.ServeHTTP(w, r)
	t.a.Equal(w.Code, code)

	return h, ps
}

// 验证指定的路径是否匹配正确的路由项，通过 code 来确定，并返回该节点的实例。
func (t *tester) matchTrue(method, path string, code int) {
	t.a.TB().Helper()

	h, _ := t.handler(method, path, code)
	t.a.NotNil(h)
}

func (t *tester) notFound(path string) {
	t.a.TB().Helper()

	hs, ps := t.tree.Route(path)
	t.a.Nil(hs).Nil(ps)
}

// 验证指定的路径返回的参数是否正确
func (t *tester) paramsTrue(method, path string, code int, params map[string]string) {
	t.a.TB().Helper()

	_, ps := t.handler(method, path, code)
	if len(params) > 0 {
		t.a.Equal(len(params), len(ps.Params))
		for k, v := range params {
			vv, found := ps.Get(k)
			t.a.True(found).Equal(vv, v)
		}
	}
}

func (t *tester) urlTrue(pattern string, params map[string]string, url string) {
	t.a.TB().Helper()

	buf := errwrap.StringBuilder{}
	err := t.tree.URL(&buf, pattern, params)
	t.a.NotError(err)
	t.a.Equal(buf.String(), url)
}

func (t *tester) urlFalse(pattern string, params map[string]string, msg string) {
	t.a.TB().Helper()

	buf := errwrap.StringBuilder{}
	err := t.tree.URL(&buf, pattern, params)
	t.a.ErrorString(err, msg)
}

// 测试 tree.methodIndex 是否正确
func (t *tester) optionsTrue(path, options string) {
	t.a.TB().Helper()

	hs, ps := t.tree.Route(path)
	t.a.Nil(ps).NotNil(hs)

	h := hs.Handler(http.MethodOptions)
	t.a.NotNil(h)

	w := httptest.NewRecorder()
	u, err := url.Parse(path)
	t.a.NotError(err).NotNil(u)
	r, err := http.NewRequest(http.MethodOptions, path, nil)
	t.a.NotError(err).NotNil(r)
	h.ServeHTTP(w, r)
	t.a.Equal(w.Code, http.StatusOK)

	t.a.Equal(w.Header().Get("Allow"), options)
}

func TestTree_AmbiguousRoute(t *testing.T) {
	a := assert.New(t, false)

	test := newTester(a, false)
	test.add(http.MethodGet, "/", 201)
	test.add(http.MethodGet, "/posts/{id}", 202)
	test.addAmbiguous("/posts/{ambiguous-id}")

	test.add(http.MethodGet, "/posts/{id}/author", 203)
	test.addAmbiguous("/posts/{-id}/author")
	test.addAmbiguous("/posts/{ambiguous-id}/author")

	test.add(http.MethodGet, "/posts-{id}-{pages}.html", 204)
	test.addAmbiguous("/posts-{id}-{ambiguous-pages}.html")
	test.addAmbiguous("/posts-{-id}-{pages}.html")
	test.addAmbiguous("/posts-{-id}-{ambiguous-pages}.html")
	test.addAmbiguous("/posts-{ambiguous-id}-{pages}.html")

	// 删除之后可以正常添加
	test.tree.Remove("/posts-{id}-{pages}.html", http.MethodGet)
	test.add(http.MethodGet, "/posts-{id}-{ambiguous-pages}.html", 203)
	test.add(http.MethodPost, "/posts-{id}-{ambiguous-pages}.html", 203)
	test.addAmbiguous("/posts-{-id}-{pages}.html")

	// 删除了 GET，依然存在 POST
	test.tree.Remove("/posts-{id}-{ambiguous-pages}.html", http.MethodGet)
	test.addAmbiguous("/posts-{-id}-{pages}.html")
	test.tree.Remove("/posts-{id}-{ambiguous-pages}.html", http.MethodPost) // POST 也删除
	test.add(http.MethodPatch, "/posts-{-id}-{pages}.html", 203)
}

func TestTree_Route(t *testing.T) {
	a := assert.New(t, false)

	test := newTester(a, false)
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
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 208) // 208 比 203 更加匹配
	test.notFound("/not-exists")

	// 仅末尾不同的路由

	test = newTester(a, false)
	test.add(http.MethodGet, "/posts/{id}", 202)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 202)

	test.add(http.MethodGet, "/posts/{id}/author", 203)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 203) // 203 比 202 更加匹配

	test.add(http.MethodGet, "/posts/{id}/{page}/author", 204)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 204) // 204 比 203 更加匹配

	test.matchTrue(http.MethodGet, "/posts/2.html/2.html", 202)
	test.add(http.MethodGet, "/posts/{id}/{page}.html", 209)    // 与 204 相结合，会将 {page} 生成一个节点，.html 生成一个节点
	test.matchTrue(http.MethodGet, "/posts/2.html/2.html", 209) // 209 比 202 更匹配

	test.matchTrue(http.MethodGet, "/posts/2.html/2/2.html", 209)
	//test.add(http.MethodGet, "/posts/{id}/{page}.html", 209)    // 与 204 相结合，会将 {page} 生成一个节点，.html 生成一个节点
	test.add(http.MethodGet, "/posts/{id}/{page}/{p2}.html", 210)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/2.html", 210) // 210 比 209 更匹配
	test.tree.Print(os.Stdout)

	// 测试 digit 和 \\d 是否正常
	test = newTester(a, true)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 201)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202)

	// 测试 digit 和 named 是否正常
	test = newTester(a, false)
	test.add(http.MethodGet, "/posts/{id}/author", 201)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202)

	// 测试 digit 和 \\d 和 named 三者顺序是否正常
	test = newTester(a, false)
	test.add(http.MethodGet, "/posts/{id}/author", 201)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 202)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 203)
	test.matchTrue(http.MethodGet, "/posts/1/author", 203)

	// 以斜框结尾，是否能正常访问
	test = newTester(a, false)
	test.add(http.MethodGet, "/posts/{id}/", 201)
	test.add(http.MethodGet, "/posts/{id}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/", 201)
	test.matchTrue(http.MethodGet, "/posts/1.html/", 201)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202)

	// 以 - 作为路径分隔符
	test = newTester(a, false)
	test.add(http.MethodGet, "/posts-{id}", 201)
	test.add(http.MethodGet, "/posts-{id}-author", 202)
	test.matchTrue(http.MethodGet, "/posts-1.html", 201)
	test.matchTrue(http.MethodGet, "/posts-1-author", 202)

	test = newTester(a, false)
	test.add(http.MethodGet, "/admin/{path}", 201)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}", 202)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile", 203)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile/{type:\\d+}", 204)
	test.matchTrue(http.MethodGet, "/admin/index.html", 201)
	test.matchTrue(http.MethodGet, "/admin/items/1", 202)
	test.matchTrue(http.MethodGet, "/admin/items/1/profile", 203)
	test.matchTrue(http.MethodGet, "/admin/items/1/profile/1", 204)

	// 测试 indexes 功能
	test = newTester(a, false)
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
	test = newTester(a, false)
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
	test = newTester(a, false)
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

	test = newTester(a, false)
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
	a := assert.New(t, false)
	test := newTester(a, false)

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
	test.paramsTrue(http.MethodGet, "/posts/", 201, map[string]string{"id": ""})
	test.paramsTrue(http.MethodGet, "/posts/1.html/author/profile/", 202, map[string]string{"id": "1.html", "action": "profile"})

	// 忽略名称捕获

	test = newTester(a, false)
	test.add(http.MethodGet, "/posts/{-id}", 201)                       // 命名
	test.add(http.MethodGet, "/posts/{-id}/author/{action}/", 202)      // 命名
	test.add(http.MethodGet, "/posts/{-id:\\d+}", 203)                  // 正则
	test.add(http.MethodGet, "/posts/{-id:\\d+}/author/{action}/", 204) // 正则

	// 正则
	test.paramsTrue(http.MethodGet, "/posts/1", 203, nil)
	test.paramsTrue(http.MethodGet, "/posts/1/author/profile/", 204, map[string]string{"action": "profile"})

	// 命名
	test.paramsTrue(http.MethodGet, "/posts/1.html", 201, nil)
	test.paramsTrue(http.MethodGet, "/posts/", 201, nil)
	test.paramsTrue(http.MethodGet, "/posts/1.html/author/profile/", 202, map[string]string{"action": "profile"})
}

func TestTreeCN(t *testing.T) {
	a := assert.New(t, false)
	test := newTester(a, false)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 201) // 命名
	test.add(http.MethodGet, "/文章/{编号}", 202)    // 中文

	test.matchTrue(http.MethodGet, "/posts/1", 201)
	test.matchTrue(http.MethodGet, "/文章/1", 202)
	test.paramsTrue(http.MethodGet, "/文章/1.html", 202, map[string]string{"编号": "1.html"})
}

func TestTree_Clean(t *testing.T) {
	a := assert.New(t, false)
	tree := New(true, syntax.NewInterceptors())

	addNode := func(p string, code int, methods ...string) {
		a.NotError(tree.Add(p, rest.BuildHandler(a, code, "", nil), methods...))
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
	a := assert.New(t, false)

	tree := New(true, syntax.NewInterceptors())
	a.NotNil(tree)

	a.NotError(tree.Add("/", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{-id}/ignore-name", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodGet, http.MethodPut, http.MethodPost))
	a.NotError(tree.Add("/posts/1/author", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", rest.BuildHandler(a, 1, "", nil), http.MethodGet))

	a.NotEmpty(tree.node.find("/posts/1/author").handlers)
	a.NotEmpty(tree.node.find("/posts/{-id}/ignore-name").handlers)
	tree.Remove("/posts/1/author", http.MethodGet)
	a.Nil(tree.node.find("/posts/1/author"))

	tree.Remove("/posts/{id}/author", http.MethodGet) // 只删除 GET
	a.NotNil(tree.node.find("/posts/{id}/author"))
	tree.Remove("/posts/{id}/author") // 删除所有请求方法
	a.Nil(tree.node.find("/posts/{id}/author"))
	tree.Remove("/posts/{id}/author") // 删除已经不存在的节点，不会报错，不发生任何事情

	// addAny

	tree = New(false, syntax.NewInterceptors())
	a.NotNil(tree)
	a.NotError(tree.Add("/path", rest.BuildHandler(a, 201, "", nil)))
	node := tree.node.find("/path")
	a.Equal(len(Methods), len(node.handlers))
	a.Equal(len(Methods), len(node.Methods()))
	a.Equal(node.Options(), strings.Join(node.Methods(), ", "))

	// head

	tree = New(true, syntax.NewInterceptors())
	a.NotNil(tree)
	a.NotError(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodGet))
	node = tree.node.find("/path")
	a.Equal(3, len(node.handlers)).
		NotNil(node.handlers[http.MethodHead]).
		NotNil(node.handlers[http.MethodOptions])

	tree.Remove("/path", http.MethodGet)
	a.Equal(0, len(node.handlers))

	// error

	tree = New(true, syntax.NewInterceptors())
	a.NotNil(tree)
	a.ErrorString(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodHead), "无法手动添加 OPTIONS/HEAD 请求方法")
	a.ErrorString(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), "NOT-SUPPORTED"), "NOT-SUPPORTED")
	a.NotError(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodDelete))
	a.ErrorString(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), http.MethodDelete), http.MethodDelete)
	tree.Remove("/path", http.MethodOptions) // remove options 不发生任何操作
	a.Equal(tree.node.find("/path").Options(), "DELETE, OPTIONS")

	a.ErrorString(tree.Add("/path/{id}/path/{id:\\d+}", rest.BuildHandler(a, 1, "", nil), http.MethodHead), "存在相同名称的路由参数")
	a.ErrorString(tree.Add("/path/{id}{id2:\\d+}", rest.BuildHandler(a, 1, "", nil), http.MethodHead), "两个命名参数不能连续出现")

	// 多层节点的删除

	tree = New(true, syntax.NewInterceptors())
	a.NotError(tree.Add("/posts", rest.BuildHandler(a, 201, "", nil), http.MethodGet, http.MethodPut))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, 202, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, 203, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author/email", rest.BuildHandler(a, 204, "", nil), http.MethodGet))

	tree.Remove("/posts/{id}") // 删除 202
	nn := tree.node.find("/posts/{id}")
	a.True(nn == nil || len(nn.handlers) == 0)

	tree.Remove("/posts/{id}/author") // 删除 203
	nn = tree.node.find("/posts/{id}/author")
	a.True(nn == nil || len(nn.handlers) == 0)

	tree.Remove("/posts/{id}/author/email") // 删除 204
	nn = tree.node.find("/posts/{id}/author/email")
	a.True(nn == nil || len(nn.handlers) == 0)
	nn = tree.node.find("/posts") // /posts 之下已经完全没有内容，所有子节点都可以删除
	a.NotEmpty(nn.handlers).Empty(nn.children)
}

func TestTree_Routes(t *testing.T) {
	a := assert.New(t, false)
	tree := New(true, syntax.NewInterceptors())
	a.NotNil(tree)

	a.NotError(tree.Add("/", rest.BuildHandler(a, http.StatusOK, "", nil), http.MethodGet))
	a.NotError(tree.Add("/posts", rest.BuildHandler(a, http.StatusOK, "", nil), http.MethodGet, http.MethodPost))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, http.StatusOK, "", nil), http.MethodGet, http.MethodPut))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, http.StatusOK, "", nil), http.MethodGet))

	routes := tree.Routes()
	a.Equal(routes, map[string][]string{
		"*":                  {http.MethodOptions},
		"/":                  {http.MethodGet, http.MethodHead, http.MethodOptions},
		"/posts":             {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost},
		"/posts/{id}":        {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut},
		"/posts/{id}/author": {http.MethodGet, http.MethodHead, http.MethodOptions},
	})
}

func TestTree_find(t *testing.T) {
	a := assert.New(t, false)
	h := rest.BuildHandler(a, http.StatusCreated, "", nil)
	tree := New(false, syntax.NewInterceptors())

	a.NotError(tree.Add("/", h, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", h, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", h, http.MethodGet))
	a.NotError(tree.Add("/posts/{-id:\\d+}/authors", h, http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", h, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", h, http.MethodGet))

	a.Equal(tree.find("/").segment.Value, "/")
	a.Equal(tree.find("/posts/{id}").segment.Value, "{id}")
	a.Equal(tree.find("/posts/{-id:\\d+}/authors").segment.Value, "{-id:\\d+}/authors")
	a.Equal(tree.find("/posts/{id}/author").segment.Value, "author")
	a.Equal(tree.find("/posts/{id}/{author:\\w+}/profile").segment.Value, "{author:\\w+}/profile")

	a.Nil(tree.find("/not-exists"))
	a.Nil(tree.find("/posts/").handlers) // 空的节点，但是有子元素。
}

func TestTree_URL(t *testing.T) {
	a := assert.New(t, false)
	test := newTester(a, true)

	// 添加路由项
	test.add(http.MethodGet, "/static", 0)                           // 静态
	test.add(http.MethodGet, "/posts/{id}", 1)                       // 命名
	test.add(http.MethodGet, "/posts/{id}/author/{action}/", 2)      // 命名
	test.add(http.MethodGet, "/posts/{id:\\d+}", 3)                  // 正则
	test.add(http.MethodGet, "/posts/{id:\\d+}/author/{action}/", 4) // 正则

	test.urlTrue("/static", nil, "/static")
	test.urlTrue("/posts/{id:\\d+}", map[string]string{"id": "100"}, "/posts/100")
	test.urlTrue("/posts/{id:\\d+}/author/{action}/", map[string]string{"id": "100", "action": "p"}, "/posts/100/author/p/")
	test.urlTrue("/posts/{id}", map[string]string{"id": "100.htm"}, "/posts/100.htm")
	test.urlTrue("/posts/{id}/author/{action}/", map[string]string{"id": "100.htm", "action": "p"}, "/posts/100.htm/author/p/")

	test.urlFalse("", nil, "并不是一条有效的注册路由项")
	test.urlFalse("/not-exists", nil, "并不是一条有效的注册路由项")
	test.urlFalse("/posts/{id}", map[string]string{"other": "other"}, "未找到参数")
	test.urlFalse("/posts/{id:\\d+}", map[string]string{"id": "xyz"}, "格式不匹配")
}
