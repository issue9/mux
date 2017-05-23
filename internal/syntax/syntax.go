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
	TypeNamed
	TypeRegexp
)

// Syntax 描述指定的字符串所表示的语法结构
type Syntax struct {
	Pattern   string   // 原始的字符串
	HasParams bool     // 是否有参数
	Wildcard  bool     // 是否为通配符模式
	Patterns  []string // 保存着对字符串处理后的结果
	Type      int      // 该语法应该被解析成的类型
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

	strs := split(pattern)
	s := &Syntax{
		Pattern:  pattern,
		Wildcard: strings.HasSuffix(pattern, "/*"),
	}
	if len(strs) == 0 ||
		(len(strs) == 1 && !isSyntax(strs[0])) {
		s.Type = TypeBasic
		return s, nil
	}

	s.Patterns = strs
	for _, v := range strs { // 判断类型
		if !isSyntax(v) {
			continue
		}

		if strings.IndexByte(v, ':') > -1 {
			s.Type = TypeRegexp
			break // 如果是正则了，则可以退出 for 了，不会再降级成其它的
		}

		s.Type = TypeNamed
	}

	names := make([]string, 0, len(s.Patterns)) // 缓存各个参数的名称，用于判断是否有重名

	// 命名参数
	if s.Type == TypeNamed {
		s.HasParams = true
		for _, str := range s.Patterns {
			lastIndex := len(str) - 1
			if str[0] == Start && str[lastIndex] == End {
				names = append(names, str[1:lastIndex])
			}
		}

		goto NAMES
	}

	// nType == typeRegexp
	for i, str := range s.Patterns {
		if !isSyntax(str) {
			continue
		}

		str = str[1 : len(str)-1] // 去掉首尾的{}符号
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

NAMES:
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
	if len(names) <= 1 {
		return -1
	}

	sort.Strings(names)
	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			return i
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
