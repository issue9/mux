// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package syntax 主要处理路由语法。
package syntax

import (
	"errors"
	"fmt"
	"strings"
)

// Type 表示路由项的类型
type Type int8

// 表示路由项的类型
const (
	TypeUnknown Type = iota
	TypeBasic
	TypeNamed
	TypeRegexp
	TypeWildcard
)

const (
	start     = '{'
	end       = '}'
	separator = ':'
)

// Segment 表示路由中最小的不可分割内容。
type Segment struct {
	Value string
	Type  Type
}

// Parse 将字符串解析成 Segment 对象数组
func Parse(str string) ([]*Segment, error) {
	ss := make([]*Segment, 0, strings.Count(str, string(start)))

	startIndex := 0
	nType := TypeBasic
	state := end // 表示当前的状态

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case start:
			if state != end {
				return nil, fmt.Errorf("不能嵌套 %s", string(start))
			}
			ss = append(ss, &Segment{
				Value: str[startIndex:i],
				Type:  nType,
			})

			startIndex = i
			state = start
			nType = TypeBasic // 记录了数据之后，重置为 TypeBasic
		case separator:
			if state != start {
				return nil, fmt.Errorf(": 只能出现在 %v %v 中间", string(start), string(end))
			}

			if i == startIndex+1 {
				return nil, errors.New("空的参数名称")
			}

			nType = TypeRegexp
			state = separator
		case end:
			if state == end {
				return nil, fmt.Errorf("%v %v 必须成对出现", string(start), string(end))
			}

			if i == startIndex+1 {
				return nil, errors.New("空的参数名称")
			}

			if state == start {
				nType = TypeNamed
			} else {
				nType = TypeRegexp
			}

			state = end
		}
	}

	if startIndex < len(str) {
		if strings.HasSuffix(str, "/*") {
			ss = append(ss, &Segment{
				Value: str[startIndex:],
				Type:  TypeWildcard,
			})
		} else {
			ss = append(ss, &Segment{
				Value: str[startIndex:],
				Type:  nType,
			})
		}
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

	startIndex := -1
	for i := 0; i < l; i++ {
		// 如果是正则，直接从 { 开始一直到结尾不再分隔，
		// 不用判断两个是否相同，不可存在两个一模一样的正则
		if s1[i] == ':' {
			return startIndex
		}

		if s1[i] == start {
			startIndex = i
		}

		if s1[i] == end {
			startIndex = -1
		}

		if s1[i] != s2[i] {
			if startIndex > -1 { // 不从命名参数中间分隔
				return startIndex
			}
			return i
		}
	}

	return l
}

// Check 检测路由项中可能包含的语法类型
func Check(pattern string) (Type, error) {
	brace := -1
	nType := TypeUnknown

	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		switch b {
		case start:
			if brace > -1 {
				return TypeUnknown, fmt.Errorf("不能嵌套 %v", start)
			}
			brace = i

			if nType != TypeRegexp {
				nType = TypeNamed
			}
		case separator:
			if brace == -1 {
				return TypeUnknown, fmt.Errorf(": 只能出现在 %v %v 中间", start, end)
			}

			if i == brace+1 {
				return TypeUnknown, errors.New("空的参数名称")
			}
			nType = TypeRegexp
		case end:
			if brace == -1 {
				return TypeUnknown, fmt.Errorf("%v %v 成对出现", start, end)
			}

			if i == brace+1 {
				return TypeUnknown, errors.New("空的参数名称")
			}
			brace = -1
		}
	}

	if strings.HasSuffix(pattern, "/*") {
		return TypeWildcard, nil
	}

	return nType, nil
}
