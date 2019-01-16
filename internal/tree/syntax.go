// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"fmt"
	"strings"
)

// SyntaxError 表示路由项的语法错误
type SyntaxError struct {
	Value   string
	Message string
}

func (err *SyntaxError) Error() string {
	return err.Message
}

// 路由项字符串中的几个特殊字符定义
const (
	nameStart       byte = '{' // 命名或是正则参数的起始字符
	nameEnd         byte = '}' // 命名或是正则参数的结束字符
	regexpSeparator byte = ':' // 正则参数中名称和正则的分隔符
)

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string("{"), "(?P<",
	string(":"), ">",
	string("}"), ")")

func isEndpoint(s string) bool {
	return s[len(s)-1] == nameEnd
}

// 获取两个字符串之间相同的前缀字符串的长度，
// 不会从 {} 中间被分开，正则表达式与之后的内容也不再分隔。
func longestPrefix(s1, s2 string) int {
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}

	startIndex := -10
	endIndex := -10
	state := nameEnd
	for i := 0; i < l; i++ {
		switch s1[i] {
		case nameStart:
			startIndex = i
			state = nameStart
		case nameEnd:
			state = nameEnd
			endIndex = i
		}

		if s1[i] != s2[i] {
			if state != nameEnd { // 不从命名参数中间分隔
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

// 获取字符串的类型。调用者需要确保 str 语法正确。
func stringType(str string) nodeType {
	typ := nodeTypeString
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case regexpSeparator:
			typ = nodeTypeRegexp
		case nameStart:
			typ = nodeTypeNamed
		case nameEnd:
			break
		}
	} // end for

	return typ
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
		state:     nameEnd,
	}
}

func (s *state) setStart(index int) *SyntaxError {
	if s.state != nameEnd {
		return &SyntaxError{Message: fmt.Sprintf("不能嵌套 %s", string(nameStart))}
	}
	if s.end+1 == index {
		return &SyntaxError{Message: "两个命名参数不能相邻"}
	}

	s.start = index
	s.state = nameStart
	return nil
}

func (s *state) setEnd(index int) *SyntaxError {
	if s.state == nameEnd {
		msg := fmt.Sprintf("%s %s 必须成对出现", string(nameStart), string(nameEnd))
		return &SyntaxError{Message: msg}
	}

	if index == s.start+1 {
		return &SyntaxError{Message: "未指定参数名称"}
	}

	if index == s.separator+1 {
		return &SyntaxError{Message: "未指定的正则表达式"}
	}

	s.state = nameEnd
	s.end = index
	return nil
}

func (s *state) setSeparator(index int) *SyntaxError {
	if s.state != nameStart {
		msg := fmt.Sprintf("字符(:)只能出现在 %s %s 中间", string(nameStart), string(nameEnd))
		return &SyntaxError{Message: msg}
	}

	if index == s.start+1 {
		return &SyntaxError{Message: "未指定参数名称"}
	}

	s.state = regexpSeparator
	s.separator = index
	return nil
}

// split 将字符串解析成字符串数组
// 以 { 为分界线进行分割。
func split(str string) ([]string, error) {
	if str == "" { // 由调用方保证不会出现此错误，所以直接 panic
		panic("参数 str 不能为空")
	}

	ss := make([]string, 0, strings.Count(str, string(nameStart))+1)

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case nameStart:
			start := state.start
			if err := state.setStart(i); err != nil {
				err.Value = str
				return nil, err
			}

			if start == i {
				continue
			}

			ss = append(ss, str[start:i])
		case regexpSeparator:
			if err := state.setSeparator(i); err != nil {
				err.Value = str
				return nil, err
			}
		case nameEnd:
			if err := state.setEnd(i); err != nil {
				err.Value = str
				return nil, err
			}
		}
	} // end for

	if state.start < len(str) {
		if state.state != nameEnd {
			msg := fmt.Sprintf("缺少 %s 字符", string(nameEnd))
			return nil, &SyntaxError{Message: msg, Value: str}
		}

		ss = append(ss, str[state.start:])
	}

	return ss, nil
}
