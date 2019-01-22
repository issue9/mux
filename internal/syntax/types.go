// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

// Type 路由项的类型
type Type int8

const (
	// TypeString 普通的字符串类型，逐字匹配，比如
	//  /users/1
	// 只能匹配 /users/1，不能匹配 /users/2
	TypeString Type = iota

	// TypeRegexp 正则表达式，比如：
	//  /users/{id:\\d+}
	// 可以匹配 /users/1、/users/2 等任意数值。
	TypeRegexp

	// TypeNamed 命名参数，相对于正则，其效率更高，当然也没有正则灵活。比如：
	//  /users/{id}
	// 可以匹配 /users/1、/users/2 和 /users/username 等非数值类型
	TypeNamed
)

// Error 表示路由项的语法错误
type Error struct {
	Value   string
	Message string
}

func (err *Error) Error() string {
	return err.Message
}

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
