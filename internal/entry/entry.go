// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"net/http"
	"regexp"
	"strings"
)

// Entry 路由项需要实现的接口
type Entry interface {
	// 匹配匹配字符串
	Pattern() string

	// 匹配程度
	//  -1 表示完全不匹配；
	//  0  表示完全匹配；
	//  >0 表示部分匹配，值越小表示匹配程度越高。
	Match(url string) int

	// 获取参数，只有正则表达式才有数据。
	Params(url string) map[string]string

	// 是否为正则表达式
	IsRegexp() bool

	// 执行该路由项的函数
	//ServeHTTP(method string, w http.ResponseWriter, r *http.Request)
	Handler(method string) http.Handler

	Add(method string, handler http.Handler) error

	Remove(method ...string) bool

	SetAllow(string)
}

//////////////// basic

// 最基本的字符串匹配路由项。
type basic struct {
	*items
	pattern string
}

func (b *basic) Pattern() string {
	return b.pattern
}

func (b *basic) IsRegexp() bool {
	return false
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

//////////////// static

// 静态文件匹配路由项，只要路径中的开头字符串与 pattern 相同，即表示匹配成功。
// 根据 match() 的返回值来确定哪个最匹配。
type static struct {
	*items
	pattern string
}

func (s *static) Pattern() string {
	return s.pattern
}

func (s *static) IsRegexp() bool {
	return false
}

func (s *static) Params(url string) map[string]string {
	return nil
}

func (s *static) Match(url string) int {
	l := len(url) - len(s.pattern)
	switch {
	case l < 0:
		return -1
	case l >= 0:
		if (s.pattern[len(s.pattern)-1] == '/') &&
			(s.pattern == url[:len(s.pattern)]) {
			return l
		}
		return -1
	} // end switch

	return -1
}

//////////////////// regexpr

// 正则表达式匹配。
type regexpr struct {
	*items
	pattern   string
	expr      *regexp.Regexp
	hasParams bool
}

func (re *regexpr) Pattern() string {
	return re.pattern
}

func (re *regexpr) IsRegexp() bool {
	return true
}

func (re *regexpr) Match(url string) int {
	// 正则匹配，没有部分匹配功能，匹配返回 0，否则返回 -1
	loc := re.expr.FindStringIndex(url)
	if loc == nil {
		return -1
	}
	if loc[0] == 0 && loc[1] == len(url) {
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
func New(pattern string, h http.Handler) Entry {
	strs := split(pattern)

	if len(strs) > 1 { // 正则路由
		p, hasParams := toPattern(strs)
		return &regexpr{
			items:     newItems(),
			pattern:   pattern,
			hasParams: hasParams,
			expr:      regexp.MustCompile(p),
		}
	}

	if pattern[len(pattern)-1] == '/' {
		return &static{
			items:   newItems(),
			pattern: pattern,
		}
	}

	return &basic{
		pattern: pattern,
		items:   newItems(),
	}
}

// 将 strs 按照顺序合并成一个正则表达式
// 返回参数正则表达式的字符串，和一个 bool 值用以表式正则中是否包含了命名匹配。
func toPattern(strs []string) (string, bool) {
	pattern := ""
	hasParams := false

	for _, v := range strs {
		lastIndex := len(v) - 1
		if v[0] != '{' || v[lastIndex] != '}' { // 普通字符串
			pattern += v
			continue
		}

		v = v[1:lastIndex] // 去掉首尾的{}符号

		index := strings.IndexByte(v, ':')
		if index < 0 { // 只存在命名，而不存在正则表达式，默认匹配[^/]
			pattern += "(?P<" + v + ">[^/]+)"
			hasParams = true
			continue
		}

		if index == 0 { // 不存在命名，但有正则表达式
			pattern += v[1:]
			continue
		}

		pattern += "(?P<" + v[:index] + ">" + v[index+1:] + ")"
		hasParams = true
	}

	return pattern, hasParams
}

// 将 str 以 { 和 } 为分隔符进行分隔。
// 符号 { 和 } 必须成对出现，且不能嵌套，否则结果是未知的。
//  /api/{id:\\d+}/users/ ==> {"/api/", "{id:\\d+}", "/users/"}
func split(str string) []string {
	ret := []string{}
	var seq byte = '{'

	for {
		if len(str) == 0 { // 没有更多字符了，结束
			break
		}

		index := strings.IndexByte(str, seq)
		if index < 0 { // 未找到分隔符，结束
			ret = append(ret, str)
			break
		}

		if seq == '}' { // 将 } 字符留在当前字符串中
			index++
		}

		if index > 0 { // 为零表示当前字符串为空，无须理会。
			ret = append(ret, str[:index])
			str = str[index:]
		}

		if seq == '{' {
			seq = '}'
		} else {
			seq = '{'
		}
	}

	return ret
}
