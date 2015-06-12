// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"regexp"
	"strings"
)

type entry struct {
	pattern   string
	expr      *regexp.Regexp
	hasParams bool
	handler   http.Handler
}

// 匹配程度
// {action:\\w+}{id:\\d+}/page
// post1/page
func (e *entry) match(url string) int {
	if e.expr == nil {
		if len(url) < len(e.pattern) {
			return -1
		}

		if e.pattern == url[:len(e.pattern)] {
			return len(url) - len(e.pattern)
		}

		return -1
	}

	if loc := e.expr.FindStringIndex(url); loc != nil {
		return len(url) - loc[1]
	}
	return -1
}

// 只有在match返回大于1的情况下，调用此函数才能返回正确结果。否则可能panic
func (e *entry) getParams(url string) map[string]string {
	if !e.hasParams {
		return nil
	}

	// 正确匹配正则表达式，则获相关的正则表达式命名变量。
	mapped := make(map[string]string)
	subexps := e.expr.SubexpNames()
	args := e.expr.FindStringSubmatch(url)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			mapped[name] = args[index]
		}
	}
	return mapped
}

func newEntry(pattern string, h http.Handler) *entry {
	strs := split(pattern)

	if len(strs) == 1 { // 静态路由
		return &entry{
			pattern: pattern,
			handler: h,
		}
	}

	pattern, hasParams := toPattern(strs)
	return &entry{
		pattern:   pattern,
		expr:      regexp.MustCompile(pattern),
		hasParams: hasParams,
		handler:   h,
	}
}

// 将strs按照顺序合并成一个正则表达式
// 返回参数正则表达式的字符串和一个bool值用以表式正则中是否包含了命名匹配。
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

// 将str以{和}为分隔符进行分隔。
// 符号{和}必须成对出现，且不能嵌套，否则结果是未知的。
func split(str string) []string {
	ret := []string{}
	var seq byte = '{'

	for {
		index := strings.IndexByte(str, seq)
		if len(str) == 0 {
			break
		}

		if index < 0 {
			ret = append(ret, str)
			break
		}

		if seq == '}' {
			index++
		}
		if index > 0 {
			ret = append(ret, str[:index])
		}
		str = str[index:]

		if seq == '{' {
			seq = '}'
		} else {
			seq = '{'
		}
	}

	return ret
}
