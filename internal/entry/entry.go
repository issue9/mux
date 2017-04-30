// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// 表示正则路由中，表达式的起止字符
const (
	RegexpStart = '{'
	RegexpEnd   = '}'
)

// 表示 Entry 接口的类型
const (
	TypeBasic = iota + 1
	TypeStatic
	TypeRegexp
)

// Entry 表示一类资源的进入点，拥有统一的路由匹配模式。
type Entry interface {
	// 返回路由的匹配字符串
	Pattern() string

	// url 与当前的匹配程度：
	//  -1 表示完全不匹配；
	//  0  表示完全匹配；
	//  >0 表示部分匹配，值越小表示匹配程度越高。
	Match(url string) int

	// 获取路由中的参数，非正则匹配返回 nil。
	Params(url string) map[string]string

	// 接口的实现类型
	Type() int

	// 获取指定请求方法对应的 http.Handler 实例，若不存在，则返回 nil。
	Handler(method string) http.Handler

	// 添加请求方法及其对应的处理函数。
	//
	// 若已经存在，则返回错误。
	// 若 method == http.MethodOptions，则可以去覆盖默认的处理方式。
	Add(handler http.Handler, methods ...string) error

	// 移除指定方法的处理池数。若 Entry 中已经没有任何 http.Handler，则返回 true
	//
	// 可以通过指定 http.MethodOptions 的方式，来强制删除 OPTIONS 请求方法的处理。
	Remove(method ...string) (empty bool)

	// 手动设置 OPTIONS 的 Allow 报头。不调用此函数，会自动根据
	// 当前的 Add 和 Remove 调整 Allow 报头，调用 SetAll() 之后，
	// 这些自动设置不再启作用。
	SetAllow(string)
}

// 最基本的字符串匹配，只能全字符串匹配。
type basic struct {
	*items
	pattern string
}

// 静态文件匹配路由项，只要路径中的开头字符串与 pattern 相同，
// 且 pattern 以 / 结尾，即表示匹配成功。根据 match() 的返回值来确定哪个最匹配。
type static struct {
	*items
	pattern string
}

// 正则表达式匹配。
type regexpr struct {
	*items
	pattern   string
	expr      *regexp.Regexp
	hasParams bool
}

func (b *basic) Pattern() string {
	return b.pattern
}

func (b *basic) Type() int {
	return TypeBasic
}

func (b *basic) Params(url string) map[string]string {
	return nil
}

func (b *basic) Match(url string) int {
	if url == b.pattern {
		return 0
	}
	return -1
}

func (s *static) Pattern() string {
	return s.pattern
}

func (s *static) Type() int {
	return TypeStatic
}

func (s *static) Params(url string) map[string]string {
	return nil
}

func (s *static) Match(url string) int {
	l := len(url) - len(s.pattern)
	if l < 0 {
		return -1
	}

	// 由 New 函数确保 s.pattern 都是以 '/' 结尾的
	if s.pattern == url[:len(s.pattern)] {
		return l
	}
	return -1
}

func (re *regexpr) Pattern() string {
	return re.pattern
}

func (re *regexpr) Type() int {
	return TypeRegexp
}

func (re *regexpr) Match(url string) int {
	loc := re.expr.FindStringIndex(url)

	if loc != nil &&
		loc[0] == 0 &&
		loc[1] == len(url) {
		return 0
	}
	return -1
}

// 将 url 与当前的表达式进行匹配，返回其命名路由参数的值。若不匹配，则返回 nil
func (re *regexpr) Params(url string) map[string]string {
	if !re.hasParams {
		return nil
	}

	// 正确匹配正则表达式，则获相关的正则表达式命名变量。
	mapped := make(map[string]string, 3)
	subexps := re.expr.SubexpNames()
	args := re.expr.FindStringSubmatch(url)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			mapped[name] = args[index]
		}
	}
	return mapped
}

// New 根据内容，生成相应的 Entry 接口实例。
//
// pattern 匹配内容。
// h 对应的 http.Handler，外层调用者确保该值不能为 nil.
func New(pattern string, h http.Handler) (Entry, error) {
	strs := split(pattern)

	if len(strs) > 1 { // 正则路由
		p, hasParams, err := toPattern(strs)
		if err != nil {
			return nil, err
		}

		expr, err := regexp.Compile(p)
		if err != nil {
			return nil, err
		}

		return &regexpr{
			items:     newItems(),
			pattern:   pattern,
			hasParams: hasParams,
			expr:      expr,
		}, nil
	}

	if pattern[len(pattern)-1] == '/' {
		return &static{
			items:   newItems(),
			pattern: pattern,
		}, nil
	}

	return &basic{
		items:   newItems(),
		pattern: pattern,
	}, nil
}

// 将 strs 按照顺序合并成一个正则表达式
// 返回参数正则表达式的字符串，和一个 bool 值用以表式正则中是否包含了命名匹配。
func toPattern(strs []string) (string, bool, error) {
	pattern := ""
	hasParams := false
	names := []string{}

	for _, v := range strs {
		lastIndex := len(v) - 1
		if v[0] != RegexpStart || v[lastIndex] != RegexpEnd { // 普通字符串
			pattern += v
			continue
		}

		v = v[1:lastIndex] // 去掉首尾的{}符号

		index := strings.IndexByte(v, ':')
		if index < 0 { // 只存在命名，而不存在正则表达式，默认匹配[^/]
			pattern += "(?P<" + v + ">[^/]+)"
			hasParams = true
			names = append(names, v)
			continue
		}

		if index == 0 { // 不存在命名，但有正则表达式
			pattern += v[1:]
			continue
		}

		pattern += "(?P<" + v[:index] + ">" + v[index+1:] + ")"
		names = append(names, v[:index])
		hasParams = true
	}

	// 检测是否存在同名参数：
	// 先按名称排序，之后只要检测相邻两个名称是否相同即可。
	if len(names) > 1 {
		sort.Strings(names)
		for i := 1; i < len(names); i++ {
			if names[i] == names[i-1] {
				return "", false, fmt.Errorf("相同的路由参数名：%v", names[i])
			}
		}
	}
	return pattern, hasParams, nil
}

// 将 str 以 { 和 } 为分隔符进行分隔。
// 符号 { 和 } 必须成对出现，且不能嵌套，否则结果是未知的。
//  /api/{id:\\d+}/users/ ==> {"/api/", "{id:\\d+}", "/users/"}
func split(str string) []string {
	ret := []string{}

	var start, end int
	for {
		if len(str) == 0 {
			return ret
		}

		start = strings.IndexByte(str, RegexpStart)
		if start < 0 { // 不存在 start
			return append(ret, str)
		}

		end = strings.IndexByte(str[start:], RegexpEnd)
		if end < 0 { // 不存在 end
			return append(ret, str)
		}
		end++
		end += start

		if start > 0 {
			ret = append(ret, str[:start])
		}

		ret = append(ret, str[start:end])
		str = str[end:]
	}
}
