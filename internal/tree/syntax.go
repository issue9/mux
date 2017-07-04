// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tree

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// 路由项字符串中的几个特殊字符定义
const (
	nameStart       byte = '{' // 命名或是正则参数的起始字符
	nameEnd         byte = '}' // 命名或是正则参数的结束字符
	regexpSeparator byte = ':' // 正则参数中名称和正则的分隔符
)

func isEndpoint(s string) bool {
	return s[len(s)-1] == nameEnd
}

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string("{"), "(?P<",
	string(":"), ">",
	string("}"), ")")

// 根据 s 内容为当前节点产生一个子节点，并返回该新节点。
func newNode(s string) (*Node, error) {
	n := &Node{
		pattern:  s,
		endpoint: isEndpoint(s),
		nodeType: stringType(s),
	}

	switch n.nodeType {
	case nodeTypeNamed:
		index := strings.IndexByte(s, nameEnd)
		if index == -1 {
			return nil, fmt.Errorf("无效的路由语法：%s", s)
		}
		n.name = s[1:index]
		n.suffix = s[index+1:]
	case nodeTypeRegexp:
		index := strings.IndexByte(s, regexpSeparator)
		if index == -1 {
			return nil, fmt.Errorf("无效的路由语法：%s", s)
		}

		expr, err := regexp.Compile(repl.Replace(s))
		if err != nil {
			return nil, err
		}
		n.expr = expr
		n.name = s[1:index]

		index = strings.IndexByte(s, nameEnd)
		if index == -1 {
			return nil, fmt.Errorf("无效的路由语法：%s", s)
		}
		n.suffix = s[index+1:]
	}

	return n, nil
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

func (s *state) setStart(index int) error {
	if s.state != nameEnd {
		return fmt.Errorf("不能嵌套 %s", string(nameStart))
	}
	if s.end+1 == index {
		return errors.New("两个命名参数不能相邻")
	}

	s.start = index
	s.state = nameStart
	return nil
}

func (s *state) setEnd(index int) error {
	if s.state == nameEnd {
		return fmt.Errorf("%s %s 必须成对出现", string(nameStart), string(nameEnd))
	}

	if index == s.start+1 {
		return errors.New("未指定参数名称")
	}

	if index == s.separator+1 {
		return errors.New("未指定的正则表达式")
	}

	s.state = nameEnd
	s.end = index
	return nil
}

func (s *state) setSeparator(index int) error {
	if s.state != nameStart {
		return fmt.Errorf("字符(:)只能出现在 %s %s 中间", string(nameStart), string(nameEnd))
	}

	if index == s.start+1 {
		return errors.New("未指定参数名称")
	}

	s.state = regexpSeparator
	s.separator = index
	return nil
}

// split 将字符串解析成字符串数组
func split(str string) ([]string, error) {
	if len(str) == 0 {
		return nil, errors.New("参数 str 不能为空")
	}

	ss := make([]string, 0, strings.Count(str, string(nameStart))+1)

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case nameStart:
			start := state.start
			if err := state.setStart(i); err != nil {
				return nil, err
			}

			if start == i {
				continue
			}

			ss = append(ss, str[start:i])
		case regexpSeparator:
			if err := state.setSeparator(i); err != nil {
				return nil, err
			}
		case nameEnd:
			if err := state.setEnd(i); err != nil {
				return nil, err
			}
		}
	} // end for

	if state.start < len(str) {
		if state.state != nameEnd {
			return nil, fmt.Errorf("缺少 %s 字符", string(nameEnd))
		}

		ss = append(ss, str[state.start:])
	}

	return ss, nil
}
