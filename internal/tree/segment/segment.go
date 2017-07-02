// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"fmt"

	"github.com/issue9/mux/params"
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
	Match(path string) (bool, string)

	// 从 path 中匹配与当前内容匹配的参数值，
	// 写入到 params 中，并返回去掉匹配内容之后的字符串。
	Params(path string, params params.Params) string

	// 将当前内容写入到 buf 中，若有参数，则参数部分内容从 params。
	URL(buf *bytes.Buffer, params map[string]string) error
}

// Equal 判断内容是否相同
func Equal(s1, s2 Segment) bool {
	return s1.Endpoint() == s2.Endpoint() &&
		s1.Value() == s2.Value() &&
		s1.Type() == s2.Type()
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
