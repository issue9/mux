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
	NameStart       byte = '{' // 命名或是正则参数的起始字符
	NameEnd         byte = '}' // 命名或是正则参数的结束字符
	RegexpSeparator byte = ':' // 正则参数中名称和正则的分隔符
)

// Parse 将字符串解析成 Segment 对象数组
func Parse(str string) ([]Segment, error) {
	if len(str) == 0 {
		return nil, errors.New("参数 str 不能为空")
	}

	ss := make([]Segment, 0, strings.Count(str, string(NameStart))+1)

	startIndex := 0
	endIndex := -10
	separatorIndex := -10

	state := NameEnd // 表示当前的状态

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case NameStart:
			if state != NameEnd {
				return nil, fmt.Errorf("不能嵌套 %s", string(NameStart))
			}
			if endIndex+1 == i {
				return nil, errors.New("两个命名参数不能相邻")
			}
			if startIndex == i {
				state = NameStart
				continue
			}

			s, err := New(str[startIndex:i])
			if err != nil {
				return nil, err
			}
			ss = append(ss, s)

			startIndex = i
			state = NameStart
		case RegexpSeparator:
			if state != NameStart {
				return nil, fmt.Errorf("字符(:)只能出现在 %s %s 中间", string(NameStart), string(NameEnd))
			}

			if i == startIndex+1 {
				return nil, errors.New("未指定参数名称")
			}

			state = RegexpSeparator
			separatorIndex = i
		case NameEnd:
			if state == NameEnd {
				return nil, fmt.Errorf("%s %s 必须成对出现", string(NameStart), string(NameEnd))
			}

			if i == startIndex+1 {
				return nil, errors.New("未指定参数名称")
			}

			if i == separatorIndex+1 {
				return nil, errors.New("未指定的正则表达式")
			}

			state = NameEnd
			endIndex = i
		}
	}

	if startIndex < len(str) {
		if state != NameEnd {
			return nil, fmt.Errorf("缺少 %s 字符", string(NameEnd))
		}

		s, err := New(str[startIndex:])
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}

	return ss, nil
}

// PrefixLen 判断两个字符串之间共同的开始内容的长度，
// 不会从{} 中间被分开，正则表达式与之后的内容也不再分隔。
func PrefixLen(s1, s2 string) int {
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}

	startIndex := -10
	endIndex := -10
	state := NameEnd
	for i := 0; i < l; i++ {
		switch s1[i] {
		case NameStart:
			startIndex = i
			state = NameStart
		case NameEnd:
			state = NameEnd
			endIndex = i
		}

		if s1[i] != s2[i] {
			if state != NameEnd { // 不从命名参数中间分隔
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
