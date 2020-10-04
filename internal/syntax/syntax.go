// SPDX-License-Identifier: MIT

// Package syntax 负责处理路由语法
package syntax

import (
	"errors"
	"fmt"
	"strings"
)

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

// 路由项字符串中的几个特殊字符定义
const (
	start     = '{' // 命名或是正则参数的起始字符
	end       = '}' // 命名或是正则参数的结束字符
	separator = ':' // 正则参数中名称和正则的分隔符
)

// 将路由语法转换成正则表达式语法，比如：
//  {id:\\d+}/author => (?P<id>\\d+)
var repl = strings.NewReplacer(string(start), "(?P<",
	string(separator), ">",
	string(end), ")")

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

// 获取字符串的类型。调用者需要确保 str 语法正确。
func getType(str string) Type {
	typ := String
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case separator:
			typ = Regexp
			break
		case start:
			typ = Named
		case end:
			break
		}
	} // end for

	return typ
}

// 将字符串解析成字符串数组
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
//  /posts/\{{id}/email ==> /posts/{, {id}/email
//  /posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func parse(str string) ([]*Segment, error) {
	if str == "" {
		return nil, errors.New("参数 str 不能为空")
	}

	ss := make([]*Segment, 0, strings.Count(str, string(start))+1)
	s := newState()

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case start:
			start := s.start
			s.setStart(i)

			if i > 0 { // i==0 表示以 { 开头
				ss = append(ss, NewSegment(str[start:i]))
			}
		case separator:
			s.setSeparator(i)
		case end:
			s.setEnd(i)
		}

		if s.err != "" {
			return nil, errors.New(s.err)
		}
	} // end for

	if s.start < len(str) {
		if s.state != end {
			return nil, fmt.Errorf("缺少 %s 字符", string(end))
		}

		ss = append(ss, NewSegment(str[s.start:]))
	}

	return ss, nil
}

// Split 将字符串解析成字符串数组
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
//  /posts/\{{id}/email ==> /posts/{, {id}/email
//  /posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func Split(str string) []*Segment {
	ss, err := parse(str)
	if err != nil {
		panic(err)
	}

	return ss
}

// IsWell 检测格式是否正确
func IsWell(str string) error {
	_, err := parse(str)
	return err
}
