// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package tree

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"
	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/internal/syntax"
	"github.com/issue9/mux/v9/internal/trace"
	"github.com/issue9/mux/v9/types"
)

type tester struct {
	tree *Tree[http.Handler]
	a    *assert.Assertion
}

func newTester(a *assert.Assertion, lock bool, trace http.Handler) *tester {
	i := syntax.NewInterceptors()
	a.NotNil(i)
	i.Add(syntax.MatchDigit, "digit")

	return &tester{
		tree: NewTestTree(a, lock, trace, i),
		a:    a,
	}
}

func testTrace(w http.ResponseWriter, r *http.Request) { trace.Trace(w, r, true) }

// 添加一条路由项。code 表示该路由项返回的报头，
// 测试路由项的 code 需要唯一，之后也是通过此值来判断其命中的路由项。
func (t *tester) add(method, pattern string, code int) {
	t.a.TB().Helper()
	t.a.NotError(t.tree.Add(pattern, rest.BuildHandler(t.a, code, "", nil), nil, method))
}

func (t *tester) addAmbiguous(pattern string) {
	t.a.TB().Helper()
	b := rest.BuildHandler(t.a, http.StatusOK, "", nil)
	t.a.ErrorString(t.tree.Add(pattern, b, nil, http.MethodGet), pattern+" 与已有的节点")
}

// 验证按照指定的 method 和 path 访问，是否会返回相同的 code 值，
// 若是，则返回该节点以及对应的参数。
func (t *tester) handler(method, path string, code int) (types.Node, http.Handler, *types.Context) {
	t.a.TB().Helper()

	ctx := types.NewContext()
	ctx.Path = path
	n, h, exists := t.tree.Handler(ctx, method)
	t.a.NotNil(n).True(exists).NotNil(h)

	w := httptest.NewRecorder()
	r := rest.NewRequest(t.a, method, path).Request()
	h.ServeHTTP(w, r)
	t.a.Equal(w.Code, code)

	return n, h, ctx
}

// 验证指定的路径是否匹配正确的路由项，通过 code 来确定，并返回该节点的实例。
func (t *tester) matchTrue(method, path string, code int, pattern string) {
	t.a.TB().Helper()

	n, h, _ := t.handler(method, path, code)
	t.a.NotNil(h).Equal(n.Pattern(), pattern)
}

func (t *tester) notFound(path string) {
	t.a.TB().Helper()

	ctx := types.NewContext()
	ctx.Path = path
	_, _, ok := t.tree.Handler(ctx, http.MethodOptions)
	t.a.False(ok)
}

// 验证指定的路径返回的参数是否正确
func (t *tester) paramsTrue(method, path string, code int, params map[string]string) {
	t.a.TB().Helper()

	_, _, ps := t.handler(method, path, code)
	if len(params) > 0 {
		t.a.Equal(len(params), ps.Count())
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

	ctx := types.NewContext()
	ctx.Path = path
	node, h, ok := t.tree.Handler(ctx, http.MethodOptions)
	t.a.True(ok).
		NotNil(node).
		NotNil(h)

	w := httptest.NewRecorder()
	u, err := url.Parse(path)
	t.a.NotError(err).NotNil(u)
	r := rest.NewRequest(t.a, http.MethodOptions, path).Request()
	h.ServeHTTP(w, r)
	t.a.Equal(w.Code, http.StatusOK)

	t.a.Equal(w.Header().Get(header.Allow), options)
}

func TestTree_AmbiguousRoute(t *testing.T) {
	a := assert.New(t, false)

	test := newTester(a, false, nil)
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

	test := newTester(a, false, nil)
	test.add(http.MethodGet, "/", 201)
	test.add(http.MethodGet, "/posts/{id}", 202)
	test.add(http.MethodGet, "/posts/{id}/author", 203)
	test.add(http.MethodGet, "/posts/1/author", 204)
	test.add(http.MethodGet, "/posts/{id:\\d+}", 205)
	test.add(http.MethodGet, "/posts/{id:\\d+}/author", 206)
	test.add(http.MethodGet, "/page/{page:\\d*}", 207) // 可选的正则节点
	test.add(http.MethodGet, "/posts/{id}/{page}/author", 208)

	test.matchTrue(http.MethodGet, "/", 201, "/")
	test.matchTrue(http.MethodGet, "/posts/1", 205, "/posts/{id:\\d+}")               // 正则
	test.matchTrue(http.MethodGet, "/posts/2", 205, "/posts/{id:\\d+}")               // 正则
	test.matchTrue(http.MethodGet, "/posts/2/author", 206, "/posts/{id:\\d+}/author") // 正则
	test.matchTrue(http.MethodGet, "/posts/1/author", 204, "/posts/1/author")         // 字符串
	test.matchTrue(http.MethodGet, "/posts/1.html", 202, "/posts/{id}")               // 命名参数
	test.matchTrue(http.MethodGet, "/posts/1.html/page", 202, "/posts/{id}")          // 命名参数
	test.matchTrue(http.MethodGet, "/posts/2.html/author", 203, "/posts/{id}/author") // 命名参数
	test.matchTrue(http.MethodGet, "/page/", 207, "/page/{page:\\d*}")
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 208, "/posts/{id}/{page}/author") // 208 比 203 更加匹配
	test.notFound("/not-exists")

	// 仅末尾不同的路由

	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/posts/{id}", 202)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 202, "/posts/{id}")

	test.add(http.MethodGet, "/posts/{id}/author", 203)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 203, "/posts/{id}/author") // 203 比 202 更加匹配

	test.add(http.MethodGet, "/posts/{id}/{page}/author", 204)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/author", 204, "/posts/{id}/{page}/author") // 204 比 203 更加匹配

	test.matchTrue(http.MethodGet, "/posts/2.html/2.html", 202, "/posts/{id}")
	test.add(http.MethodGet, "/posts/{id}/{page}.html", 209)                               // 与 204 相结合，会将 {page} 生成一个节点，.html 生成一个节点
	test.matchTrue(http.MethodGet, "/posts/2.html/2.html", 209, "/posts/{id}/{page}.html") // 209 比 202 更匹配

	test.matchTrue(http.MethodGet, "/posts/2.html/2/2.html", 209, "/posts/{id}/{page}.html")
	test.add(http.MethodGet, "/posts/{id}/{page}/{p2}.html", 210)
	test.matchTrue(http.MethodGet, "/posts/2.html/2/2.html", 210, "/posts/{id}/{page}/{p2}.html") // 210 比 209 更匹配

	// 测试 digit 和 \\d 是否正常
	test = newTester(a, true, nil)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 201)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/author", 202, "/posts/{id:digit}/author")

	// 测试 digit 和 named 是否正常
	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 202)
	test.add(http.MethodGet, "/posts/{id}/author", 201)
	test.matchTrue(http.MethodGet, "/posts/xxx/author", 201, "/posts/{id}/author")
	test.matchTrue(http.MethodGet, "/posts/1/author", 202, "/posts/{id:digit}/author")

	// 测试 digit 和 \\d 和 named 三者顺序是否正常
	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/posts/{id}/author", 201)
	test.add(http.MethodGet, "/posts/{id:\\d}/author", 202)
	test.add(http.MethodGet, "/posts/{id:digit}/author", 203)
	test.matchTrue(http.MethodGet, "/posts/1/author", 203, "/posts/{id:digit}/author")

	// 以斜框结尾，是否能正常访问
	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/posts/{id}/", 201)
	test.add(http.MethodGet, "/posts/{id}/author", 202)
	test.matchTrue(http.MethodGet, "/posts/1/", 201, "/posts/{id}/")
	test.matchTrue(http.MethodGet, "/posts/1.html/", 201, "/posts/{id}/")
	test.matchTrue(http.MethodGet, "/posts/1/author", 202, "/posts/{id}/author")

	// 以 - 作为路径分隔符
	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/posts-{id}", 201)
	test.add(http.MethodGet, "/posts-{id}-author", 202)
	test.matchTrue(http.MethodGet, "/posts-1.html", 201, "/posts-{id}")
	test.matchTrue(http.MethodGet, "/posts-1-author", 202, "/posts-{id}-author")

	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/admin/{path}", 201)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}", 202)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile", 203)
	test.add(http.MethodGet, "/admin/items/{id:\\d+}/profile/{type:\\d+}", 204)
	test.matchTrue(http.MethodGet, "/admin/index.html", 201, "/admin/{path}")
	test.matchTrue(http.MethodGet, "/admin/items/1", 202, "/admin/items/{id:\\d+}")
	test.matchTrue(http.MethodGet, "/admin/items/1/profile", 203, "/admin/items/{id:\\d+}/profile")
	test.matchTrue(http.MethodGet, "/admin/items/1/profile/1", 204, "/admin/items/{id:\\d+}/profile/{type:\\d+}")

	// 测试 indexes 功能
	test = newTester(a, false, nil)
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
	test.matchTrue(http.MethodGet, "/admin/5index.html", 220, "/admin/{id}")

	// 测试非英文字符 indexes 功能
	test = newTester(a, false, nil)
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
	test.matchTrue(http.MethodGet, "/中文/5index.html", 220, "/中文/{id}")

	// 测试非英文字符 indexes 功能，汉字中的相同部分会被提取到上一级。
	test = newTester(a, false, nil)
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
	test.matchTrue(http.MethodGet, "/中文/5index.html", 220, "/中文/{id}")
	test.optionsTrue("/中文/5index.html", "GET, HEAD, OPTIONS")

	// "OPTIONS trace=false

	test = newTester(a, false, nil)
	test.add(http.MethodGet, "/admin/1", 201)
	test.matchTrue(http.MethodGet, "/admin/1", 201, "/admin/1")
	test.optionsTrue("/admin/1", "GET, HEAD, OPTIONS")
	test.optionsTrue("", "GET, OPTIONS")
	test.optionsTrue("*", "GET, OPTIONS")

	// 动态添加其它方法，会改变 OPTIONS 值
	test.add(http.MethodPost, "/admin/1", 201)
	test.optionsTrue("", "GET, OPTIONS, POST")

	// OPTIONS trace=true

	test = newTester(a, false, http.HandlerFunc(testTrace))
	test.add(http.MethodGet, "/admin/1", 201)
	test.matchTrue(http.MethodGet, "/admin/1", 201, "/admin/1")
	test.optionsTrue("/admin/1", "GET, HEAD, OPTIONS, TRACE")
	test.optionsTrue("", "GET, OPTIONS, TRACE")
	test.optionsTrue("*", "GET, OPTIONS, TRACE")

	// 动态添加其它方法，会改变 OPTIONS 值
	test.add(http.MethodPost, "/admin/1", 201)
	test.optionsTrue("", "GET, OPTIONS, POST, TRACE")
}

func TestTree_Params(t *testing.T) {
	a := assert.New(t, false)
	test := newTester(a, false, nil)

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

	test = newTester(a, false, nil)
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
	test := newTester(a, false, nil)

	// 添加路由项
	test.add(http.MethodGet, "/posts/{id}", 201) // 命名
	test.add(http.MethodGet, "/文章/{编号}", 202)    // 中文

	test.matchTrue(http.MethodGet, "/posts/1", 201, "/posts/{id}")
	test.matchTrue(http.MethodGet, "/文章/1", 202, "/文章/{编号}")
	test.paramsTrue(http.MethodGet, "/文章/1.html", 202, map[string]string{"编号": "1.html"})
}

func TestTree_Clean(t *testing.T) {
	a := assert.New(t, false)
	tree := NewTestTree(a, true, nil, syntax.NewInterceptors())

	addNode := func(p string, code int, methods ...string) {
		a.NotError(tree.Add(p, rest.BuildHandler(a, code, "", nil), nil, methods...))
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

	tree := NewTestTree(a, true, nil, syntax.NewInterceptors())

	a.NotError(tree.Add("/", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{-id}/ignore-name", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodGet, http.MethodPut, http.MethodPost))
	a.NotError(tree.Add("/posts/1/author", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", rest.BuildHandler(a, 1, "", nil), nil, http.MethodGet))

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

	tree = NewTestTree(a, false, nil, syntax.NewInterceptors())
	a.NotError(tree.Add("/path", rest.BuildHandler(a, 201, "", nil), nil))
	node := tree.node.find("/path")
	a.Equal(len(Methods), len(node.handlers))    // 多了 methodNotAllowed，但是 trace 并不保存在 handlers 中
	a.Equal(len(Methods)-1, len(node.Methods())) // methodNotAllowed 和 trace 并不记入 node.Methods()
	a.Equal(node.AllowHeader(), strings.Join(node.Methods(), ", "))

	// OPTIONS

	tree = NewTestTree(a, true, nil, syntax.NewInterceptors())
	a.NotError(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodGet))
	node = tree.node.find("/path")
	a.Equal(4, len(node.handlers)).
		NotNil(node.handlers[http.MethodOptions])

	tree.Remove("/path", http.MethodGet)
	a.Equal(0, len(node.handlers))

	// error

	tree = NewTestTree(a, true, nil, syntax.NewInterceptors())
	a.ErrorString(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, "NOT-SUPPORTED"), "NOT-SUPPORTED")
	a.NotError(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodDelete))
	a.ErrorString(tree.Add("/path", rest.BuildHandler(a, http.StatusAccepted, "", nil), nil, http.MethodDelete), http.MethodDelete)
	tree.Remove("/path", http.MethodOptions) // remove options 不发生任何操作
	a.Equal(tree.node.find("/path").AllowHeader(), "DELETE, OPTIONS")

	a.ErrorString(tree.Add("/path/{id}/path/{id:\\d+}", rest.BuildHandler(a, 1, "", nil), nil, http.MethodHead), "存在相同名称的路由参数")
	a.ErrorString(tree.Add("/path/{id}{id2:\\d+}", rest.BuildHandler(a, 1, "", nil), nil, http.MethodHead), "两个命名参数不能连续出现")

	// 多层节点的删除

	tree = NewTestTree(a, true, nil, syntax.NewInterceptors())
	a.NotError(tree.Add("/posts", rest.BuildHandler(a, 201, "", nil), nil, http.MethodGet, http.MethodPut))
	a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, 202, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, 203, "", nil), nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author/email", rest.BuildHandler(a, 204, "", nil), nil, http.MethodGet))

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

	t.Run("trace=false", func(t *testing.T) {
		tree := NewTestTree(a, true, nil, syntax.NewInterceptors())

		a.NotError(tree.Add("/", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet))
		a.NotError(tree.Add("/posts", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet, http.MethodPost))
		a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet, http.MethodPut))
		a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet))

		routes := tree.Routes()
		a.Equal(routes, map[string][]string{
			"*":                  {http.MethodOptions},
			"/":                  {http.MethodGet, http.MethodHead, http.MethodOptions},
			"/posts":             {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost},
			"/posts/{id}":        {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut},
			"/posts/{id}/author": {http.MethodGet, http.MethodHead, http.MethodOptions},
		})
	})

	t.Run("trace=true", func(t *testing.T) {
		tree := NewTestTree(a, true, http.HandlerFunc(testTrace), syntax.NewInterceptors())

		a.NotError(tree.Add("/", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet))
		a.NotError(tree.Add("/posts", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet, http.MethodPost))
		a.NotError(tree.Add("/posts/{id}", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet, http.MethodPut))
		a.NotError(tree.Add("/posts/{id}/author", rest.BuildHandler(a, http.StatusOK, "", nil), nil, http.MethodGet))

		routes := tree.Routes()
		a.Equal(routes, map[string][]string{
			"*":                  {http.MethodOptions, http.MethodTrace},
			"/":                  {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
			"/posts":             {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodTrace},
			"/posts/{id}":        {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodTrace},
			"/posts/{id}/author": {http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
		})
	})
}

func TestTree_find(t *testing.T) {
	a := assert.New(t, false)
	h := rest.BuildHandler(a, http.StatusCreated, "", nil)
	tree := NewTestTree(a, false, nil, syntax.NewInterceptors())

	a.NotError(tree.Add("/", h, nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}", h, nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/author", h, nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{-id:\\d+}/authors", h, nil, http.MethodGet))
	a.NotError(tree.Add("/posts/1/author", h, nil, http.MethodGet))
	a.NotError(tree.Add("/posts/{id}/{author:\\w+}/profile", h, nil, http.MethodGet))

	a.Equal(tree.Find("/").segment.Value, "/")
	a.Equal(tree.Find("/posts/{id}").segment.Value, "{id}")
	a.Equal(tree.Find("/posts/{-id:\\d+}/authors").segment.Value, "{-id:\\d+}/authors")
	a.Equal(tree.Find("/posts/{id}/author").segment.Value, "author")
	a.Equal(tree.Find("/posts/{id}/{author:\\w+}/profile").segment.Value, "{author:\\w+}/profile")

	a.Nil(tree.Find("/not-exists"))
	a.Nil(tree.Find("/posts/").handlers) // 空的节点，但是有子元素。
}

func TestTree_URL(t *testing.T) {
	a := assert.New(t, false)
	test := newTester(a, true, nil)

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

func TestTree_match(t *testing.T) {
	a := assert.New(t, false)
	tree := NewTestTree(a, false, nil, syntax.NewInterceptors())

	// path1，主动调用 WriteHeader

	a.NotError(tree.Add("/path1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("h1", "h1")
		w.WriteHeader(http.StatusAccepted)

		_, err := w.Write([]byte("get"))
		a.NotError(err)
		_, err = w.Write([]byte("get"))
		a.NotError(err)
	}), nil, http.MethodGet))

	ctx := types.NewContext()
	ctx.Path = "/path1"
	node, h, ok := tree.Handler(ctx, http.MethodOptions)
	a.True(ok).
		NotNil(node).
		NotNil(h).
		Zero(ctx.Count())

	w := httptest.NewRecorder()
	r := rest.NewRequest(a, http.MethodOptions, "/path1").Request()
	h.ServeHTTP(w, r)
	a.Empty(w.Header().Get("h1")).
		Equal(w.Header().Get(header.Allow), "GET, HEAD, OPTIONS").
		Empty(w.Body.String())

	// path2，不主动调用 WriteHeader

	a.NotError(tree.Add("/path2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("h1", "h2")

		_, err := w.Write([]byte("get"))
		a.NotError(err)
		_, err = w.Write([]byte("get"))
		a.NotError(err)
	}), nil, http.MethodGet))

	ctx = types.NewContext()
	ctx.Path = "/path2"
	node, h, ok = tree.Handler(ctx, http.MethodOptions)
	a.True(ok).
		NotNil(node).
		NotNil(h).
		Zero(ctx.Count())

	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/path2").Request()
	h.ServeHTTP(w, r)
	a.Empty(w.Header().Get("h1")).
		Equal(w.Header().Get(header.Allow), "GET, HEAD, OPTIONS").
		Empty(w.Body.String())
}

func TestTree_ApplyMiddleware(t *testing.T) {
	a := assert.New(t, false)

	tree := NewTestTree(a, false, nil, syntax.NewInterceptors())

	err := tree.Add("/m", rest.BuildHandler(a, http.StatusOK, "/m/", nil), nil, http.MethodGet)
	a.NotError(err)

	err = tree.Add("/m/path", rest.BuildHandler(a, http.StatusOK, "/m/path/", nil), nil, http.MethodGet)
	a.NotError(err)

	tree.ApplyMiddleware(BuildTestMiddleware(a, "m1"), BuildTestMiddleware(a, "m2"))

	newCtx := func(path string) *types.Context {
		ctx := types.NewContext()
		ctx.Path = path
		return ctx
	}

	// GET /m
	_, f, exists := tree.Handler(newCtx("/m"), http.MethodGet)
	a.True(exists).NotNil(f)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/m").Request()
	f.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "/m/m1m2")

	// HEAD /m
	_, f, exists = tree.Handler(newCtx("/m"), http.MethodHead)
	a.True(exists).NotNil(f)
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodHead, "/m").Request()
	f.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "/m/m1m2")

	// GET /m/path
	_, f, exists = tree.Handler(newCtx("/m/path"), http.MethodGet)
	a.True(exists).NotNil(f)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/m/path").Request()
	f.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "/m/path/m1m2")

	// OPTIONS /m/path
	_, f, exists = tree.Handler(newCtx("/m/path"), http.MethodOptions)
	a.True(exists).NotNil(f)
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/m/path").Request()
	f.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "m1m2"). // m1m2 由中间件产生
						Equal(w.Header().Get(header.Allow), "GET, HEAD, OPTIONS")

	// DELETE /m/path  method not allowed
	_, f, exists = tree.Handler(newCtx("/m/path"), http.MethodDelete)
	a.False(exists).NotNil(f)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/m/path").Request()
	f.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "m1m2"). // m1m2 由中间件产生
						Equal(w.Result().StatusCode, http.StatusMethodNotAllowed)

	// DELETE /not-exists  not found
	_, f, exists = tree.Handler(newCtx("/not-exists"), http.MethodDelete)
	a.False(exists).NotNil(f)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/m/path").Request()
	f.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "404 page not found\nm1m2"). // m1m2 由中间件产生
								Equal(w.Result().StatusCode, http.StatusNotFound)
}

func TestTree_Handler(t *testing.T) {
	a := assert.New(t, false)
	tree := NewTestTree(a, false, nil, syntax.NewInterceptors())

	tree.Add("/path1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/path1"))
	}), nil, http.MethodDelete, http.MethodGet)

	// path 不存在
	ctx := types.NewContext()
	ctx.Path = "/path"
	n, h, exists := tree.Handler(ctx, http.MethodDelete)
	a.False(exists).NotNil(h).Nil(n)

	// method 不存在
	ctx = types.NewContext()
	ctx.Path = "/path1"
	n, h, exists = tree.Handler(ctx, http.MethodPut)
	a.False(exists).NotNil(h).NotNil(n)

	ctx = types.NewContext()
	ctx.Path = "/path1"
	n, h, exists = tree.Handler(ctx, http.MethodHead)
	a.True(exists).NotNil(h).NotNil(n)
}
