// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"regexp"
	"strings"
)

type entryer interface {
	match(url string) (int, map[string]string)
}

// /post/1
type staticEntry string

func (e staticEntry) match(url string) (int, map[string]string) {
	if len(url) < len(e) {
		return -1, nil
	}

	if string(e) == url[:len(e)] {
		return len(url) - len(e), nil
	}

	return -1, nil
}

type regexpEntry struct {
	expr      *regexp.Regexp
	hasParams bool
}

// {action:\\w+}{id:\\d+}/page
// post1/page
func (e regexpEntry) match(url string) (int, map[string]string) {
	loc := e.expr.FindStringIndex(url)
	if loc == nil {
		return -1, nil
	}

	r := len(url) - loc[1]
	if !e.hasParams {
		return r, nil
	}

	// 正确匹配正则表达式，则获相关的正则表达式命名变量。
	mapped := make(map[string]string)
	subexps := e.expr.SubexpNames()
	args := e.expr.FindStringSubmatch(url)
	for index, name := range subexps {
		if len(name) > 0 {
			mapped[name] = args[index]
		}
	}
	return r, mapped
}

func newEntry(pattern string) entryer {
	strs := split(pattern)

	if len(strs) == 1 { // 静态路由
		return staticEntry(pattern)
	}

	pattern, hasParams := toPattern(strs)
	return &regexpEntry{
		expr:      regexp.MustCompile(pattern),
		hasParams: hasParams,
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
