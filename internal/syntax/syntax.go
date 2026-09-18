// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package syntax 负责处理路由语法
package syntax

import (
	"errors"
	"fmt"
	"iter"
	"strings"

	"github.com/issue9/errwrap"
)

// Type 路由项节点的类型
type Type int8

const (
	// String 普通的字符串类型，逐字匹配，比如
	//  /users/1
	// 只能匹配 /users/1，不能匹配 /users/2
	String Type = iota

	// Interceptor 拦截器
	//
	// 这是正则和命名参数的特例，其优先级比两都都要高。
	Interceptor

	// Regexp 正则表达式，比如：
	//  {id:\\d+}/abc
	// 可以匹配 /users/1、/users/2 等任意数值。
	Regexp

	// Named 命名参数，相对于正则，其效率更高，当然也没有正则灵活。比如：
	//  {id}/abc
	// 可以匹配 /users/1、/users/2 和 /users/username 等非数值类型
	Named
)

// 路由项字符串中的几个特殊字符定义
const (
	startByte     = '{' // 命名或是正则参数的起始字符
	endByte       = '}' // 命名或是正则参数的结束字符
	separatorByte = ':' // 正则参数中名称和正则的分隔符
	ignoreByte    = '-' // 忽略名称的前缀
)

func (t Type) String() string {
	switch t {
	case Named:
		return "named"
	case Interceptor:
		return "interceptor"
	case Regexp:
		return "regexp"
	case String:
		return "string"
	default:
		panic("不存在的类型")
	}
}

// URL 将 ps 中的参数填入 pattern
//
// 如果存在于 pattern，不存在于 ps，将出错；
// 如果存在于 ps，不存在于 pattern 则是可以的。
func (i *Interceptors) URL(buf *errwrap.StringBuilder, pattern string, ps map[string]string) error {
	if pattern == "" {
		return nil
	}

	segments, err := i.Split(pattern)
	if err != nil {
		return err
	}

	for _, seg := range segments {
		if seg.Type == String {
			buf.WString(seg.Value)
			continue
		}

		val, found := ps[seg.Name]
		if !found {
			return fmt.Errorf("未找到参数 %s 的值", seg.Name)
		}
		buf.WString(val).WString(seg.Suffix)
	}

	return nil
}

// Split 将字符串解析成 [Segment] 数组
//
// 以 { 为分界线进行分割。比如
//
//	/posts/{id}/email ==> /posts/, {id}/email
//	/posts/\{{id}/email ==> /posts/{, {id}/email
//	/posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func (i *Interceptors) Split(str string) ([]*Segment, error) {
	if str == "" {
		return nil, errors.New("参数 str 不能为空")
	}

	ss, size := splitString(str)
	segments := make([]*Segment, 0, size)
	var lastFlag bool
	names := make(map[string]int, size)

	for s := range ss {
		if lastFlag && s[0] == startByte {
			return nil, fmt.Errorf("两个命名参数不能连续出现：%s", str)
		}
		lastFlag = s[len(s)-1] == endByte

		seg, err := i.NewSegment(s)
		if err != nil {
			return nil, err
		}

		if seg.Type != String {
			if names[seg.Name] > 0 {
				return nil, fmt.Errorf("存在相同名称的路由参数：%s", seg.Name)
			}
			names[seg.Name]++
		}

		segments = append(segments, seg)
	}

	return segments, nil
}

func splitString(str string) (iter.Seq[string], int) {
	size := strings.Count(str, string(startByte)) + 1
	var end int

	return func(yield func(s string) bool) {
		for {
			start := strings.IndexByte(str[end:], startByte)
			if start == -1 {
				yield(str)
				break
			} else if start > 0 {
				if yield(str[:start+end]) {
					str = str[start+end:]
				} else {
					break
				}
			}

			end = strings.IndexByte(str, endByte)
			if end == -1 {
				yield(str)
				break
			}
		}
	}, size
}
