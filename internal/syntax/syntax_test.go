// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestType_String(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(Named.String(), "named")
	a.Equal(Interceptor.String(), "interceptor")
	a.Equal(Regexp.String(), "regexp")
	a.Equal(String.String(), "string")
	a.Panic(func() {
		_ = (Type(5)).String()
	})
}

func TestInterceptor_Split(t *testing.T) {
	a := assert.New(t, false)
	i := NewInterceptors()
	a.NotNil(i)

	test := func(str string, isError bool, ss ...string) {
		s, err := i.Split(str)

		if isError {
			a.Error(err)
			return
		}

		a.NotError(err).Equal(len(s), len(ss))
		for index, str := range ss {
			seg, err := i.NewSegment(str)
			a.NotError(err).NotNil(seg)

			item := s[index]
			a.Equal(seg.Value, item.Value).
				Equal(seg.Name, item.Name).
				Equal(seg.Endpoint, item.Endpoint).
				Equal(seg.Suffix, item.Suffix)
		}
	}

	test("/", false, "/")

	test("/posts/1", false, "/posts/1")
	test("/posts/}/author", false, "/posts/}/author")
	test("/posts/:id/author", false, "/posts/:id/author")

	test("{action}/1", false, "{action}/1")
	test("{act/ion}/1", false, "{act/ion}/1") // 名称中包含非常规则字符
	test("{中文}/1", false, "{中文}/1")           // 名称中包含中文

	// 以命名参数开头的
	test("/{action}", false, "/", "{action}")

	// : 出现在 {} 之外
	test("{中文}/:1", false, "{中文}/:1")

	// 以通配符结尾
	test("/posts/{id}", false, "/posts/", "{id}")

	test("/posts/{id}/author/profile", false, "/posts/", "{id}/author/profile")
	test("/posts/{id:}", false, "/posts/", "{id:}")
	test("/posts/{id}/{author", false, "/posts/", "{id}/", "{author")

	// 以命名参数结尾的
	test("/posts/{id}/author", false, "/posts/", "{id}/author")

	test("/posts/{id:digit}/author", false, "/posts/", "{id:digit}/author")

	// 命名参数及通配符
	test("/posts/{id}/page/{page}", false, "/posts/", "{id}/page/", "{page}")
	test("/posts/{id}/page/{page:digit}", false, "/posts/", "{id}/page/", "{page:digit}")

	// 正则
	test("/posts/{id:\\d+}", false, "/posts/", "{id:\\d+}")

	// 正则，命名参数
	test("/posts/{id:\\d+}/page/{page}", false, "/posts/", "{id:\\d+}/page/", "{page}")

	// 正则不捕获参数
	test("/posts/{-id:\\d+}/page/{page}", false, "/posts/", "{-id:\\d+}/page/", "{page}")

	// 命名参数，不捕获参数值
	test("/posts/{-id:}/page/{page}", false, "/posts/", "{-id:}/page/", "{page}")

	// 一些错误格式
	test("/posts/{{id:\\d+}/author", true)
	test("/posts/{:\\d+}/author", true)
	test("/posts/{}/author", true)
	test("/posts/{id}{page}/", true) // 连续的参数
	test("/posts/{id}-{id}/", true)  // 同名
	test("", true)
}

func TestSplitString(t *testing.T) {
	a := assert.New(t, false)
	test := func(input string, output ...string) {
		ss := splitString(input)
		a.Equal(ss, output)
	}

	test("/", "/")

	test("/posts/1", "/posts/1")

	test("/{path", "/", "{path")
	test("{action}/1", "{action}/1")
	test("{act/ion}/1", "{act/ion}/1") // 名称中包含非常规则字符
	test("{中文}/1", "{中文}/1")           // 名称中包含中文

	// 以命名参数开头的
	test("/{action}", "/", "{action}")

	// : 出现在 {} 之外
	test("{中文}/:1", "{中文}/:1")

	// 以通配符结尾
	test("/posts/{id}", "/posts/", "{id}")

	test("/posts/{id}/author/profile", "/posts/", "{id}/author/profile")
	test("/posts/{id:}", "/posts/", "{id:}")

	// 以命名参数结尾的
	test("/posts/{id}/author", "/posts/", "{id}/author")

	test("/posts/{id:digit}/author", "/posts/", "{id:digit}/author")

	// 命名参数及通配符
	test("/posts/{id}/page/{page}", "/posts/", "{id}/page/", "{page}")
	test("/posts/{id}/page/{page:digit}", "/posts/", "{id}/page/", "{page:digit}")

	// 正则
	test("/posts/{id:\\d+}", "/posts/", "{id:\\d+}")

	// 正则，命名参数
	test("/posts/{id:\\d+}/page/{page}", "/posts/", "{id:\\d+}/page/", "{page}")

	test("/posts/{{id:\\d+}/author", "/posts/", "{{id:\\d+}/author")
}

func TestTree_URL(t *testing.T) {
	a := assert.New(t, false)
	i := NewInterceptors()
	a.NotNil(i)

	data := []*struct {
		pattern string
		ps      map[string]string
		output  string
		err     bool
	}{
		{
			pattern: "",
			output:  "",
		},
		{
			pattern: "/static",
			output:  "/static",
		},
		{
			pattern: "/posts/{id}",
			ps:      map[string]string{"id": "100"},
			output:  "/posts/100",
		},
		{
			pattern: "/posts/{id",
			ps:      map[string]string{"id": "100"},
			output:  "/posts/{id",
		},
		{
			pattern: "/posts/{id:}",
			ps:      map[string]string{"id": "100"},
			output:  "/posts/100",
		},
		{ // ignoreName=true
			pattern: "/posts/{-id:}",
			ps:      map[string]string{"id": "100"},
			output:  "/posts/100",
		},
		{
			pattern: "/posts/{id}:",
			ps:      map[string]string{"id": "100"},
			output:  "/posts/100:",
		},
		{
			pattern: "/posts/{id:\\d+}",
			ps:      map[string]string{"id": "100"},
			output:  "/posts/100",
		},
		{
			pattern: "/posts/{id:\\d+}/author/{page}/",
			ps:      map[string]string{"id": "100", "page": "200"},
			output:  "/posts/100/author/200/",
		},
		{ // ignoreName=true
			pattern: "/posts/{-id:\\d+}/author/{page}/",
			ps:      map[string]string{"id": "100", "page": "200"},
			output:  "/posts/100/author/200/",
		},
		{
			pattern: "/posts/{编号}/作者/{page}/",
			ps:      map[string]string{"编号": "100", "page": "200"},
			output:  "/posts/100/作者/200/",
		},
		{ // 参数未指定，直接判断全部节点为 String
			pattern: "/posts/{id:\\d+}",
			output:  "/posts/{id:\\d+}",
		},

		{ // pattern 格式错误
			pattern: "/posts/{{id:\\d+}",
			ps:      map[string]string{"id": "1"},
			err:     true,
		},
	}

	for _, item := range data {
		output, err := i.URL(item.pattern, item.ps)
		if item.err {
			a.Error(err, "解析 %s 未出现预期的错误", item.pattern).
				Empty(item.output)
		} else {
			a.NotError(err, "解析 %s 出现错误: %s", item.pattern, err).
				Equal(output, item.output)
		}
	}
}
