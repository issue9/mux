// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import "fmt"

type state struct {
	start     int
	end       int
	separator int
	state     byte
	err       string // 错误信息
}

func newState() *state {
	return &state{
		start:     0,
		end:       -10,
		separator: -10,
		state:     end,
	}
}

func (s *state) setStart(index int) {
	if s.err != "" {
		return
	}

	if s.state != end {
		s.err = fmt.Sprintf("不能嵌套 %s", string(start))
		return
	}

	if s.end+1 == index {
		s.err = "两个命名参数不能相邻"
		return
	}

	s.start = index
	s.state = start
}

func (s *state) setEnd(index int) {
	if s.err != "" {
		return
	}

	if s.state == end {
		s.err = fmt.Sprintf("%s %s 必须成对出现", string(start), string(end))
		return
	}

	if index == s.start+1 {
		s.err = "未指定参数名称"
		return
	}

	if index == s.separator+1 {
		s.err = "未指定的正则表达式"
		return
	}

	s.state = end
	s.end = index
}

func (s *state) setSeparator(index int) {
	if s.err != "" {
		return
	}

	if s.state != start {
		s.err = fmt.Sprintf("字符(:)只能出现在 %s %s 中间", string(start), string(end))
		return
	}

	if index == s.start+1 {
		s.err = "未指定参数名称"
		return
	}

	s.state = separator
	s.separator = index
}
