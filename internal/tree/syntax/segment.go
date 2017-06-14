// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

// Segment 表示路由中的分段内容。
type Segment struct {
	Value    string
	Type     Type
	Endpoint bool // 是否为终点
}

// NewSegment 将字符串转换为一个 Segment 实例。
// 调用者需要确保 str 语法正确。
func NewSegment(str string) *Segment {
	return &Segment{
		Value:    str,
		Type:     stringType(str),
		Endpoint: str[len(str)-1] == NameEnd,
	}
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
