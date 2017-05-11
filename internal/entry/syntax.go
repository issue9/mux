// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"fmt"
	"sort"
	"strings"
)

// 表示正则路由中，表达式的起止字符
const (
	syntaxStart = '{'
	syntaxEnd   = '}'
)

// 表示当前路由项的类型，同时会被用于 Entry.priority()
const (
	typeUnknown = iota
	typeBasic
	typeRegexp
	typeNamed

	// 通配符版本的，优先级比没有通配符的都要低。
	typeBasicWithWildcard
	typeRegexpWithWildcard
	typeNamedWithWildcard
)

// 描述指定的字符串所表示的语法结构
type syntax struct {
	hasParams bool     // 是否有参数
	patterns  []string // 保存着对字符串处理后的结果
	nType     int      // 该语法应该被解析成的类型
}

// 判断 str 是一个合法的语法结构还是普通的字符串
func isSyntax(str string) bool {
	return str[0] == syntaxStart && str[len(str)-1] == syntaxEnd
}

// 对字符串进行分析，判断其类型，以及是否包含参数
func parse(pattern string) (*syntax, error) {
	names := []string{} // 缓存各个参数的名称，用于判断是否有重名

	strs := split(pattern)
	s := &syntax{
		patterns: make([]string, 0, len(strs)),
	}
	if len(strs) == 0 {
		s.nType = typeBasic
		return s, nil
	}

	if len(strs) == 1 && !isSyntax(strs[0]) {
		s.patterns = append(s.patterns, strs[0])
		s.nType = typeBasic
		return s, nil
	}

	// 判断类型
	for _, v := range strs {
		s.patterns = append(s.patterns, v)

		if !isSyntax(v) { // 普通字符串
			continue
		}

		// 只存在命名，而不存在正则表达式
		if index := strings.IndexByte(v, ':'); index < 0 {
			if s.nType != typeRegexp {
				s.nType = typeNamed
			}
		} else {
			s.nType = typeRegexp
		}
	}

	// 命名参数
	if s.nType == typeNamed {
		for _, str := range s.patterns {
			lastIndex := len(str) - 1
			if str[0] == syntaxStart && str[lastIndex] == syntaxEnd {
				names = append(names, str[1:lastIndex])
				s.hasParams = true
			}
		}

		goto DUP
	}

	// nType == typeRegexp
	for i, str := range s.patterns {
		lastIndex := len(str) - 1
		if !isSyntax(str) {
			continue
		}

		str = str[1:lastIndex] // 去掉首尾的{}符号

		index := strings.IndexByte(str, ':')
		if index < 0 { // 只存在命名，而不存在正则表达式，默认匹配[^/]
			s.patterns[i] = "(?P<" + str + ">[^/]+)"
			s.hasParams = true
			names = append(names, str)
			continue
		}

		if index == 0 { // 不存在命名，但有正则表达式
			s.patterns[i] = str[1:]
			continue
		}

		s.patterns[i] = "(?P<" + str[:index] + ">" + str[index+1:] + ")"
		s.hasParams = true
		names = append(names, str[:index])
	}

DUP:
	if index := duplicateName(names); index >= 0 {
		return nil, fmt.Errorf("相同的路由参数名：%v", names[index])
	}
	return s, nil
}

func duplicateName(names []string) int {
	// 检测是否存在同名参数：
	// 先按名称排序，之后只要检测相邻两个名称是否相同即可。
	if len(names) > 1 {
		sort.Strings(names)
		for i := 1; i < len(names); i++ {
			if names[i] == names[i-1] {
				return i
			}
		}
	}

	return -1
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

		start = strings.IndexByte(str, syntaxStart)
		if start < 0 { // 不存在 start
			return append(ret, str)
		}

		end = strings.IndexByte(str[start:], syntaxEnd)
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
