// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"strings"
)

type state struct {
	start     int
	end       int
	separator int
	state     byte
	err       string // 错误信息
	value     string // 需要被解析的字符串
}

func newState(v string) *state {
	s := &state{
		start:     0,
		end:       -10,
		separator: -10,
		state:     end,
		value:     v,
	}

	if v == "" {
		s.err = "值不能为空"
	}

	return s
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

// Split 将字符串解析成字符串数组。
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
//  /posts/\{{id}/email ==> /posts/{, {id}/email
//  /posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func (s *state) split() []*Segment {
	ss := make([]*Segment, 0, strings.Count(s.value, string(start))+1)

	for i := 0; i < len(s.value); i++ {
		switch s.value[i] {
		case start:
			start := s.start
			s.setStart(i)

			if i == 0 { // 以 { 开头
				continue
			}

			ss = append(ss, NewSegment(s.value[start:i]))
		case separator:
			s.setSeparator(i)
		case end:
			s.setEnd(i)
		}

		if s.err != "" {
			panic(s.err)
		}
	} // end for

	if s.err != "" {
		panic(s.err)
	}

	if s.start < len(s.value) {
		if s.state != end {
			panic(fmt.Sprintf("缺少 %s 字符", string(end)))
		}

		ss = append(ss, NewSegment(s.value[s.start:]))
	}

	return ss
}

// 检测格式是否正确
func (s *state) isWell() string {
	for i := 0; i < len(s.value); i++ {
		switch s.value[i] {
		case start:
			s.setStart(i)

			if i == 0 { // 以 { 开头
				continue
			}
		case separator:
			s.setSeparator(i)
		case end:
			s.setEnd(i)
		}

		if s.err != "" {
			return s.err
		}
	} // end for

	if s.err != "" {
		return s.err
	}

	if s.start < len(s.value) {
		if s.state != end {
			return fmt.Sprintf("缺少 %s 字符", string(end))
		}
	}

	return ""
}
