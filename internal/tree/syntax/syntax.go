// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package syntax 处理路由语法。
package syntax

import (
	"errors"
	"fmt"
	"strings"
)

// Type 表示路由项的类型
type Type int8

// 表示路由项的类型，同时也表示节点的匹配优先级，值越小优先级越高。
const (
	TypeBasic Type = iota + 1
	TypeNamedBasic
	TypeNamed
	TypeRegexp
	TypeWildcard
)

// 路由项字符串中的几个特殊字符定义
const (
	NameStart       byte = '{' // 包含命名参数的起始字符
	NameEnd         byte = '}' // 包含命名参数的结束字符
	RegexpSeparator byte = ':' // 名称和正则的分隔符
	Wildcard        byte = '*' // 通配符
)

// Segment 表示路由中最小的不可分割内容。
type Segment struct {
	Value string
	Type  Type
}

// Parse 将字符串解析成 Segment 对象数组
func Parse(str string) ([]*Segment, error) {
	ss := make([]*Segment, 0, strings.Count(str, string(NameStart)))

	startIndex := 0
	endIndex := -1
	separatorIndex := -1

	nType := TypeBasic
	state := NameEnd // 表示当前的状态
	isLast := len(str) - 1

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case Wildcard:
			if i < isLast {
				return nil, fmt.Errorf("%s 只能出现在结尾", string(Wildcard))
			}

			if endIndex+1 == i {
				return nil, fmt.Errorf("不能同时出现 %v %v", string(NameEnd), string(Wildcard))
			}

			ss = append(ss, &Segment{
				Value: str[startIndex:i],
				Type:  nType,
			}, &Segment{
				Value: "*",
				Type:  TypeWildcard,
			})

			return ss, nil
		case NameStart:
			if state != NameEnd {
				return nil, fmt.Errorf("不能嵌套 %s", string(NameStart))
			}
			if endIndex+1 == i {
				return nil, errors.New("两个命名参数不能相邻")
			}

			if nType == TypeNamed {
				nType = TypeNamedBasic
			}

			ss = append(ss, &Segment{
				Value: str[startIndex:i],
				Type:  nType,
			})

			startIndex = i
			state = NameStart
			nType = TypeBasic // 记录了数据之后，重置为 TypeBasic
		case RegexpSeparator:
			if state != NameStart {
				return nil, fmt.Errorf(": 只能出现在 %v %v 中间", string(NameStart), string(NameEnd))
			}

			if i == startIndex+1 {
				return nil, errors.New("空的参数名称")
			}

			nType = TypeRegexp
			state = RegexpSeparator
			separatorIndex = i
		case NameEnd:
			if state == NameEnd {
				return nil, fmt.Errorf("%v %v 必须成对出现", string(NameStart), string(NameEnd))
			}

			if i == startIndex+1 {
				return nil, errors.New("空的参数名称")
			}

			if i == separatorIndex+1 {
				return nil, fmt.Errorf("无效的字符 %v", string(RegexpSeparator))
			}

			if state == NameStart {
				nType = TypeNamed
			} else {
				nType = TypeRegexp
			}

			state = NameEnd
			endIndex = i
		}
	}

	if startIndex < len(str) {
		if nType == TypeNamed && str[len(str)-1] != NameEnd {
			nType = TypeNamedBasic
		}

		ss = append(ss, &Segment{
			Value: str[startIndex:],
			Type:  nType,
		})
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

		if s1[i] == NameStart {
			startIndex = i
		}

		if s1[i] == NameEnd {
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
