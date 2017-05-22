// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package syntax 提供对路由字符串语法的分析功能。
package syntax

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// 路由中允许的最大参数数量
const maxParamsSize = 255

// 正则语法的起止字符
const (
	Start = '{'
	End   = '}'
)

// 表示当前路由项的类型，同时会被用于 Entry.Priority()
const (
	TypeUnknown = iota
	TypeBasic
	TypeRegexp
	TypeNamed
)

// Syntax 描述指定的字符串所表示的语法结构
type Syntax struct {
	Pattern   string   // 原始的字符串
	HasParams bool     // 是否有参数
	Wildcard  bool     // 是否为通配符模式
	Patterns  []string // 保存着对字符串处理后的结果
	Type      int      // 该语法应该被解析成的类型
}

// IsWildcard 当前字符串是否为通配符模式
func IsWildcard(pattern string) bool {
	return strings.HasSuffix(pattern, "/*")
}

// 判断 str 是一个合法的语法结构还是普通的字符串
func isSyntax(str string) bool {
	return str[0] == Start && str[len(str)-1] == End
}

// New 根据 pattern 生成一个 *Syntax 对象
func New(pattern string) (*Syntax, error) {
	if len(pattern) == 0 || pattern[0] != '/' {
		return nil, errors.New("参数 pattern 必须以 / 开头")
	}

	names := []string{} // 缓存各个参数的名称，用于判断是否有重名

	strs := split(pattern)
	s := &Syntax{
		Pattern:  pattern,
		Patterns: make([]string, 0, len(strs)),
		Wildcard: IsWildcard(pattern),
	}
	if len(strs) == 0 ||
		len(strs) == 1 && !isSyntax(strs[0]) {
		s.Type = TypeBasic
		return s, nil
	}

	// 判断类型
	for _, v := range strs {
		s.Patterns = append(s.Patterns, v)

		if !isSyntax(v) { // 普通字符串
			continue
		}

		// 只存在命名，而不存在正则表达式
		if index := strings.IndexByte(v, ':'); index < 0 {
			if s.Type != TypeRegexp {
				s.Type = TypeNamed
			}
		} else {
			s.Type = TypeRegexp
		}
	}

	// 命名参数
	if s.Type == TypeNamed {
		for _, str := range s.Patterns {
			lastIndex := len(str) - 1
			if str[0] == Start && str[lastIndex] == End {
				names = append(names, str[1:lastIndex])
				s.HasParams = true
			}
		}

		goto DUP
	}

	// nType == typeRegexp
	for i, str := range s.Patterns {
		lastIndex := len(str) - 1
		if !isSyntax(str) {
			continue
		}

		str = str[1:lastIndex] // 去掉首尾的{}符号

		index := strings.IndexByte(str, ':')
		if index < 0 { // 只存在命名，而不存在正则表达式，默认匹配[^/]
			s.Patterns[i] = "(?P<" + str + ">[^/]+)"
			s.HasParams = true
			names = append(names, str)
			continue
		}

		if index == 0 { // 不存在命名，但有正则表达式
			s.Patterns[i] = str[1:]
			continue
		}

		s.Patterns[i] = "(?P<" + str[:index] + ">" + str[index+1:] + ")"
		s.HasParams = true
		names = append(names, str[:index])
	}

DUP:
	if index := duplicateName(names); index >= 0 {
		return nil, fmt.Errorf("相同的路由参数名：%v", names[index])
	}

	if len(names) >= maxParamsSize {
		return nil, fmt.Errorf("路径参数最多只能 %d 个，当前数量 %d", maxParamsSize, len(names))
	}

	return s, nil
}

func duplicateName(names []string) int {
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

		start = strings.IndexByte(str, Start)
		if start < 0 { // 不存在 start
			return append(ret, str)
		}

		end = strings.IndexByte(str[start:], End)
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
