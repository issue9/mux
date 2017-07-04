// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"errors"
	"fmt"
	"strings"
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

// Split 将字符串解析成字符串数组
func Split(str string) ([]string, error) {
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
