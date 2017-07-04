// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import "strings"

// 路由项字符串中的几个特殊字符定义
const (
	NameStart       byte = '{' // 命名或是正则参数的起始字符
	NameEnd         byte = '}' // 命名或是正则参数的结束字符
	RegexpSeparator byte = ':' // 正则参数中名称和正则的分隔符
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

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string("{"), "(?P<",
	string(":"), ">",
	string("}"), ")")

// Regexp 将 str 转换成一条正常表达式字符串
func Regexp(str string) string {
	return repl.Replace(str)
}

// IsEndpoint 判断字符串是否可作为终点。
func IsEndpoint(str string) bool {
	return str[len(str)-1] == NameEnd
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

// StringType 获取字符串的类型。调用者需要确保 str 语法正确。
func StringType(str string) Type {
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
