// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

// Type 路由项的类型
type Type int8

// 表示路由项的类型
const (
	TypeString Type = iota
	TypeRegexp
	TypeNamed
)

// 仅上面的 trace 用到
func (t Type) String() string {
	switch t {
	case TypeNamed:
		return "named"
	case TypeRegexp:
		return "regexp"
	case TypeString:
		return "string"
	default:
		return "<unknown>"
	}
}

// GetType 获取字符串的类型。调用者需要确保 str 语法正确。
func GetType(str string) Type {
	typ := TypeString
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case Separator:
			typ = TypeRegexp
		case Start:
			typ = TypeNamed
		case End:
			break
		}
	} // end for

	return typ
}
