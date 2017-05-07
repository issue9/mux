// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// 表示正则路由中，表达式的起止字符
const (
	syntaxStart = '{'
	syntaxEnd   = '}'
)

var ErrIsNotRegexp = errors.New("delet")

type syntax struct {
	hasParams bool
	patterns  []string // 保存着对字符串处理后的结果
	nType     int
}

func (s *syntax) add(patterns ...string) {
	s.patterns = append(s.patterns, patterns...)
}

// Parse 分析 pattern，如果可能的话就将其转换成正则表达式字符串。
// 具体语法可参照根目录的文档。
//
// 返回参数正则表达式的字符串，和一个 bool 值用以表式正则中是否包含了命名匹配。
func Parse(pattern string) (string, bool, error) {
	hasParams := false
	names := []string{}

	strs := split(pattern)
	if len(strs) == 1 && (strs[0][0] != syntaxStart || strs[0][len(strs[0])-1] != syntaxEnd) {
		return "", false, ErrIsNotRegexp
	}

	pattern = pattern[:0]
	for _, v := range strs {
		lastIndex := len(v) - 1
		if v[0] != syntaxStart || v[lastIndex] != syntaxEnd { // 普通字符串
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

// 对字符串进行分析，判断其类型，以及是否包含参数
func parse(pattern string) (*syntax, error) {
	names := []string{} // 缓存各个参数的名称，用于判断是否有重名

	strs := split(pattern)
	s := &syntax{
		patterns: make([]string, 0, len(strs)),
	}
	if len(strs) == 1 && (strs[0][0] != syntaxStart || strs[0][len(strs[0])-1] != syntaxEnd) {
		s.add(strs[0])
		s.nType = TypeBasic
		return s, nil
	}

	// 判断类型
	for _, v := range strs {
		s.add(v)

		lastIndex := len(v) - 1
		if v[0] != syntaxStart || v[lastIndex] != syntaxEnd { // 普通字符串
			continue
		}

		// 只存在命名，而不存在正则表达式
		if index := strings.IndexByte(v, ':'); index < 0 {
			if s.nType != TypeRegexp {
				s.nType = TypeNamed
			}
		} else {
			s.nType = TypeRegexp
		}
	}

	// 命名参数
	if s.nType == TypeNamed {
		for _, str := range s.patterns {
			lastIndex := len(str) - 1
			if str[0] == syntaxStart && str[lastIndex] == syntaxEnd {
				names = append(names, str[1:lastIndex])
				s.hasParams = true
			}
		}

		goto DUP
	}

	// nType == TypeRegexp
	for i, str := range s.patterns {
		lastIndex := len(str) - 1
		if str[0] != syntaxStart || str[lastIndex] != syntaxEnd {
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
			s.add(str[1:])
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
