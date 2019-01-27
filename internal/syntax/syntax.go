// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package syntax 负责处理路由语法
package syntax

import "strings"

// Type 路由项的类型
type Type int8

const (
	// String 普通的字符串类型，逐字匹配，比如
	//  /users/1
	// 只能匹配 /users/1，不能匹配 /users/2
	String Type = iota

	// Regexp 正则表达式，比如：
	//  /users/{id:\\d+}
	// 可以匹配 /users/1、/users/2 等任意数值。
	Regexp

	// Named 命名参数，相对于正则，其效率更高，当然也没有正则灵活。比如：
	//  /users/{id}
	// 可以匹配 /users/1、/users/2 和 /users/username 等非数值类型
	Named
)

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string(Start), "(?P<",
	string(Separator), ">",
	string(End), ")")

func isEndpoint(s string) bool {
	return s[len(s)-1] == End
}

// 仅上面的 trace 用到
func (t Type) String() string {
	switch t {
	case Named:
		return "named"
	case Regexp:
		return "regexp"
	case String:
		return "string"
	default:
		panic("不存在的类型")
	}
}

// getType 获取字符串的类型。调用者需要确保 str 语法正确。
func getType(str string) Type {
	typ := String
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case Separator:
			typ = Regexp
		case Start:
			typ = Named
		case End:
			break
		}
	} // end for

	return typ
}
