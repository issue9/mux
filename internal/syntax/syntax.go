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
	Start     byte = '{' // 命名或是正则参数的起始字符
	End       byte = '}' // 命名或是正则参数的结束字符
	Separator byte = ':' // 正则参数中名称和正则的分隔符
)

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string(Start), "(?P<",
	string(Separator), ">",
	string(End), ")")

// ToRegexp 将内容转换成正则表达式
func ToRegexp(expr string) *regexp.Regexp {
	return regexp.MustCompile(repl.Replace(expr))
}

// IsEndpoint 是否为最终结点
//
// 在非字符路由项中，如果以 {path} 等结尾，
// 可以匹配任意剩余字符，此函数用于判断是否为该功能的结点。
func IsEndpoint(s string) bool {
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

type state struct {
	start     int
	end       int
	separator int
	state     byte
}

func newState() *state {
	return &state{
		start:     0,
		end:       -10,
		separator: -10,
		state:     End,
	}
}

func (s *state) setStart(index int) *Error {
	if s.state != End {
		return &Error{Message: fmt.Sprintf("不能嵌套 %s", string(Start))}
	}
	if s.end+1 == index {
		return &Error{Message: "两个命名参数不能相邻"}
	}

	s.start = index
	s.state = Start
	return nil
}

func (s *state) setEnd(index int) *Error {
	if s.state == End {
		msg := fmt.Sprintf("%s %s 必须成对出现", string(Start), string(End))
		return &Error{Message: msg}
	}

	if index == s.start+1 {
		return &Error{Message: "未指定参数名称"}
	}

	if index == s.separator+1 {
		return &Error{Message: "未指定的正则表达式"}
	}

	s.state = End
	s.end = index
	return nil
}

func (s *state) setSeparator(index int) *Error {
	if s.state != Start {
		msg := fmt.Sprintf("字符(:)只能出现在 %s %s 中间", string(Start), string(End))
		return &Error{Message: msg}
	}

	if index == s.start+1 {
		return &Error{Message: "未指定参数名称"}
	}

	s.state = Separator
	s.separator = index
	return nil
}

// Split 将字符串解析成字符串数组。
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
func Split(str string) ([]string, error) {
	if str == "" { // 由调用方保证不会出现此错误，所以直接 panic
		panic("参数 str 不能为空")
	}

	ss := make([]string, 0, strings.Count(str, string(Start))+1)

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case Start:
			start := state.start
			if err := state.setStart(i); err != nil {
				err.Value = str
				return nil, err
			}

			if start == i {
				continue
			}

			ss = append(ss, str[start:i])
		case Separator:
			if err := state.setSeparator(i); err != nil {
				err.Value = str
				return nil, err
			}
		case End:
			if err := state.setEnd(i); err != nil {
				err.Value = str
				return nil, err
			}
		}
	} // end for

	if state.start < len(str) {
		if state.state != End {
			msg := fmt.Sprintf("缺少 %s 字符", string(End))
			return nil, &Error{Message: msg, Value: str}
		}

		ss = append(ss, str[state.start:])
	}

	return ss, nil
}
