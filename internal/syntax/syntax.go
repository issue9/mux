// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package syntax 负责处理路由语法
package syntax

import (
	"fmt"
	"regexp"
	"strings"
)

// 路由项字符串中的几个特殊字符定义
const (
	Start     = '{'  // 命名或是正则参数的起始字符
	End       = '}'  // 命名或是正则参数的结束字符
	Separator = ':'  // 正则参数中名称和正则的分隔符
	Escape    = '\\' // 路由项中的转义字符
)

// Segment 路由项被拆分之后的值
type Segment struct {
	Value string
	Type  Type

	// 是否为最终结点
	//
	// 在非字符路由项中，如果以 {path} 等结尾，可以匹配任意剩余字符。
	Endpoint bool
}

// NewSegment 声明新的 Segment 变量
func NewSegment(val string) *Segment {
	seg := &Segment{
		Value: val,
		Type:  GetType(val),
	}

	if seg.Type != String {
		seg.Endpoint = isEndpoint(val)
	}

	return seg
}

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string(Start), "(?P<",
	string(Separator), ">",
	string(End), ")")

// ToRegexp 将内容转换成正则表达式
func ToRegexp(expr string) *regexp.Regexp {
	return regexp.MustCompile(repl.Replace(expr))
}

func isEndpoint(s string) bool {
	return s[len(s)-1] == End
}

// LongestPrefix 获取两个字符串之间相同的前缀字符串的长度，
// 不会从 {} 中间被分开，正则表达式与之后的内容也不再分隔。
func LongestPrefix(s1, s2 string) int {
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}

	startIndex := -10
	endIndex := -10
	state := End
	for i := 0; i < l; i++ {
		switch s1[i] {
		case Start:
			startIndex = i
			state = Start
		case End:
			state = End
			endIndex = i
		}

		if s1[i] != s2[i] {
			if state != End { // 不从命名参数中间分隔
				return startIndex
			}
			if endIndex == i { // 命名参数之后必须要有一个或以上的普通字符
				return startIndex
			}
			return i
		}
	} // end for

	if endIndex == l-1 {
		return startIndex
	}

	return l
}

// Split 将字符串解析成字符串数组。
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
//  /posts/\{{id}/email ==> /posts/{, {id}/email
//  /posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func Split(str string) []*Segment {
	if str == "" {
		panic("参数 str 不能为空")
	}

	ss := make([]*Segment, 0, strings.Count(str, string(Start))+1)

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case Start:
			start := state.start
			state.setStart(i)

			if i == 0 { // 以 { 开头
				continue
			}

			ss = append(ss, NewSegment(str[start:i]))
		case Separator:
			state.setSeparator(i)
		case End:
			state.setEnd(i)
		}

		if state.err != "" {
			panic(state.err)
		}
	} // end for

	if state.err != "" {
		panic(state.err)
	}

	if state.start < len(str) {
		if state.state != End {
			panic(fmt.Sprintf("缺少 %s 字符", string(End)))
		}

		ss = append(ss, NewSegment(str[state.start:]))
	}

	return ss
}

// IsWell 检测格式是否正确
func IsWell(str string) string {
	if str == "" {
		return "参数 str 不能为空"
	}

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case Start:
			state.setStart(i)

			if i == 0 { // 以 { 开头
				continue
			}
		case Separator:
			state.setSeparator(i)
		case End:
			state.setEnd(i)
		}

		if state.err != "" {
			return state.err
		}
	} // end for

	if state.err != "" {
		return state.err
	}

	if state.start < len(str) {
		if state.state != End {
			return fmt.Sprintf("缺少 %s 字符", string(End))
		}
	}

	return ""
}
