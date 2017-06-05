// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package syntax 处理路由字符串语法。
package syntax

import (
	"errors"
	"fmt"
	"strings"
)

// Type 表示路由项的类型
type Type int8

// 表示路由项的类型。
// 同时也会被用于表示节点的匹配优先级，值越小优先级越高。
const (
	TypeBasic Type = iota + 10
	TypeRegexp
	TypeNamed
	//TypeWildcard // 通配符类型，即以命名参数结尾
)

// 路由项字符串中的几个特殊字符定义
const (
	NameStart       byte = '{' // 包含命名参数的起始字符
	NameEnd         byte = '}' // 包含命名参数的结束字符
	RegexpSeparator byte = ':' // 名称和正则的分隔符
)

// Segment 表示路由中最小的不可分割内容。
type Segment struct {
	Value    string
	Type     Type
	Endpoint bool // 是否为终点
}

// Parse 将字符串解析成 Segment 对象数组
func Parse(str string) ([]*Segment, error) {
	ss := make([]*Segment, 0, strings.Count(str, string(NameStart)))

	startIndex := 0
	endIndex := -1
	separatorIndex := -1

	nType := TypeBasic
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

			ss = append(ss, &Segment{
				Value: str[startIndex:i],
				Type:  nType,
			})

			startIndex = i
			state = NameStart
			nType = TypeBasic // 记录了数据之后，重置为 TypeBasic
		case RegexpSeparator:
			if state != NameStart {
				return nil, fmt.Errorf("字符(:)只能出现在 %v %v 中间", string(NameStart), string(NameEnd))
			}

			if i == startIndex+1 {
				return nil, errors.New("未指定参数名称")
			}

			state = RegexpSeparator
			separatorIndex = i
		case NameEnd:
			if state == NameEnd {
				return nil, fmt.Errorf("%v %v 必须成对出现", string(NameStart), string(NameEnd))
			}

			if i == startIndex+1 {
				return nil, errors.New("未指定参数名称")
			}

			if i == separatorIndex+1 {
				return nil, errors.New("未指定的正则表达式")
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
		if state != NameEnd {
			return nil, fmt.Errorf("缺少 %s 字符", string(NameEnd))
		}

		// 最后一个节点是命名节点，则转换成通配符模式
		ss = append(ss, &Segment{
			Value:    str[startIndex:],
			Type:     nType,
			Endpoint: str[len(str)-1] == NameEnd,
		})
	}

	return ss, nil
}

var repl = strings.NewReplacer(string(NameStart), "(?P<",
	string(RegexpSeparator), ">",
	string(NameEnd), ")")

// Regexp 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
// 需要保证 pattern 的语法正确，此处不再做检测。
func Regexp(pattern string) string {
	return repl.Replace(pattern)
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

// NewSegment 将字符串声明为一个 Segment 实例
func NewSegment(str string) *Segment {
	return &Segment{
		Value: str,
		Type:  stringType(str),
	}
}

// 获取字符串的类型。调用者需要确保 str 语法正确。
func stringType(str string) Type {
	typ := TypeBasic

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case RegexpSeparator:
			typ = TypeRegexp
		case NameStart:
			typ = TypeNamed
		case NameEnd:
			break
		}
	} // end for

	return typ
}
