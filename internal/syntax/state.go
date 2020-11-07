// SPDX-License-Identifier: MIT

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
	s := &state{}
	s.reset()

	return s
}

func (s *state) reset() {
	s.start = 0
	s.end = -10
	s.separator = -10
	s.state = endByte
	s.err = ""
}

func (s *state) setStart(index int) {
	if s.state != endByte {
		s.err = fmt.Sprintf("不能嵌套 %s", string(startByte))
		return
	}

	if s.end+1 == index {
		s.err = "两个命名参数不能相邻"
		return
	}

	s.start = index
	s.state = startByte
}

func (s *state) setEnd(index int) {
	if s.state == endByte {
		s.err = fmt.Sprintf("%s %s 必须成对出现", string(startByte), string(endByte))
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

	s.state = endByte
	s.end = index
}

func (s *state) setSeparator(index int) {
	if s.state != startByte {
		s.err = fmt.Sprintf("字符(%s)只能出现在 %s %s 中间", string(separatorByte), string(startByte), string(endByte))
		return
	}

	if index == s.start+1 {
		s.err = "未指定参数名称"
		return
	}

	s.state = separatorByte
	s.separator = index
}
