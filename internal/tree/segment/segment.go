// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package segment 处理路由字符串语法。
package segment

import (
	"bytes"
	"fmt"

	"github.com/issue9/mux/params"
)

// 路由项字符串中的几个特殊字符定义
const (
	nameStart       byte = '{' // 命名或是正则参数的起始字符
	nameEnd         byte = '}' // 命名或是正则参数的结束字符
	regexpSeparator byte = ':' // 正则参数中名称和正则的分隔符
)

// Type 表示当前 Segment 的类型
type Type int8

// 表示 Segment 的类型。
// 同时也表示各个类型在被匹配时的优先级顺序。
const (
	TypeString Type = iota
	TypeRegexp
	TypeNamed
)

// Segment 表示路由中的分段内容。
type Segment interface {
	// 当前内容的类型。
	Type() Type

	// 当前内容的值
	Value() string

	// 用于表示当前是否为终点，仅对 Type 为 TypeRegexp 和 TypeNamed 有用。
	// 此值为 true，该节点的优先级会比同类型的节点低，以便优先对比其它非最终节点。
	Endpoint() bool

	// 将当前内容与 path 进行匹配，若成功匹配，
	// 则返回 true 和去掉匹配内容之后的字符串。
	Match(path string, params params.Params) (bool, string)

	// 将当前内容写入到 buf 中，若有参数，则参数部分内容从 params。
	URL(buf *bytes.Buffer, params map[string]string) error

	// 从 params 中删除当前的内容对应的参数。
	DeleteParams(params params.Params)
}

// Equal 判断内容是否相同
func Equal(s1, s2 Segment) bool {
	return s1.Endpoint() == s2.Endpoint() &&
		s1.Value() == s2.Value() // 值相同，类型值肯定相同
}

// IsEndpoint 判断字符串是否可作为终点。
func IsEndpoint(str string) bool {
	return str[len(str)-1] == nameEnd
}

// New 将字符串转换为一个 Segment 实例。
// 调用者需要确保 str 语法正确。
func New(s string) (Segment, error) {
	typ := stringType(s)
	switch typ {
	case TypeNamed:
		return newNamed(s)
	case TypeString:
		return str(s), nil
	case TypeRegexp:
		return newReg(s)
	}
	return nil, fmt.Errorf("无效的节点类型 %d", typ)
}

// 获取字符串的类型。调用者需要确保 str 语法正确。
func stringType(str string) Type {
	typ := TypeString

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case regexpSeparator:
			typ = TypeRegexp
		case nameStart:
			typ = TypeNamed
		case nameEnd:
			break
		}
	} // end for

	return typ
}

// LongestPrefix 获取两个字符串之间相同的前缀字符串的长度，
// 不会从 {} 中间被分开，正则表达式与之后的内容也不再分隔。
func LongestPrefix(s1, s2 string) int {
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
