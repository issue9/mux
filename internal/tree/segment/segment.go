// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"fmt"
)

// Type 表示当前 Segment 的类型
type Type int8

// 表示 Segment 的类型。
const (
	TypeString Type = iota * 10
	TypeRegexp
	TypeNamed
)

// Segment 表示路由中的分段内容。
type Segment interface {
	Type() Type

	Pattern() string

	// 用于表示当前是否为终点，仅对 Type 为 TypeRegexp 和 TypeNamed 有用。
	// 此值为 true，该节点的优先级会比同类型的节点低，以便优先对比其它非最终节点。
	Endpoint() bool

	Match(path string) (bool, string)

	Params(path string, params map[string]string) string

	URL(buf *bytes.Buffer, params map[string]string) error
}

// New 将字符串转换为一个 Segment 实例。
// 调用者需要确保 str 语法正确。
func New(str string) (Segment, error) {
	typ := stringType(str)
	switch typ {
	case TypeNamed:
		return newNamed(str)
	case TypeString:
		return newStr(str)
	case TypeRegexp:
		return newReg(str)
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
