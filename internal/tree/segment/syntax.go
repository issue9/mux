// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package segment 处理路由字符串语法。
package segment

import (
	"errors"
	"fmt"
	"strings"
)

// 路由项字符串中的几个特殊字符定义
const (
	nameStart       byte = '{' // 命名或是正则参数的起始字符
	nameEnd         byte = '}' // 命名或是正则参数的结束字符
	regexpSeparator byte = ':' // 正则参数中名称和正则的分隔符
)

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

// Parse 将字符串解析成 Segment 对象数组
func Parse(str string) ([]Segment, error) {
	if len(str) == 0 {
		return nil, errors.New("参数 str 不能为空")
	}

	ss := make([]Segment, 0, strings.Count(str, string(nameStart))+1)
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

			s, err := New(str[start:i])
			if err != nil {
				return nil, err
			}
			ss = append(ss, s)
		case regexpSeparator:
			if err := state.setSeparator(i); err != nil {
				return nil, err
			}
		case nameEnd:
			if err := state.setEnd(i); err != nil {
				return nil, err
			}
		}
	}

	if state.start < len(str) {
		if state.state != nameEnd {
			return nil, fmt.Errorf("缺少 %s 字符", string(nameEnd))
		}

		s, err := New(str[state.start:])
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}

	return ss, nil
}

// PrefixLen 判断两个字符串之间共同的开始内容的长度，
// 不会从 {} 中间被分开，正则表达式与之后的内容也不再分隔。
func PrefixLen(s1, s2 string) int {
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
