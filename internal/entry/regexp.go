// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// 表示正则路由中，表达式的起止字符
const (
	RegexpStart = '{'
	RegexpEnd   = '}'
)

// Regexp 正则表达式匹配。
type Regexp struct {
	*items
	pattern   string
	expr      *regexp.Regexp
	hasParams bool
}

// Pattern 匹配字符串
func (re *Regexp) Pattern() string {
	return re.pattern
}

// Type 类型
func (re *Regexp) Type() int {
	return TypeRegexp
}

// Match url 与当前的匹配程序
func (re *Regexp) Match(url string) int {
	loc := re.expr.FindStringIndex(url)

	if loc != nil &&
		loc[0] == 0 &&
		loc[1] == len(url) {
		return 0
	}
	return -1
}

// Params 将 url 与当前的表达式进行匹配，返回其命名路由参数的值。若不匹配，则返回 nil
func (re *Regexp) Params(url string) map[string]string {
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
