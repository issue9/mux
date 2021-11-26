// SPDX-License-Identifier: MIT

// Package syntax 负责处理路由语法
package syntax

import (
	"errors"
	"fmt"
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
// 如果 pattern 中存在，但是不存在于 ps，将出错，
// 但是如果只存在于 ps，但是不存在于 pattern 是可以的。
//
// 不能将 URL 作为判断 pattern 是否合规的方法，在 ps 为空时， 将直接返回 pattern。
func (i *Interceptors) URL(buf *errwrap.StringBuilder, pattern string, ps map[string]string) error {
	if pattern == "" {
		return nil
	}

	segs, err := i.Split(pattern)
	if err != nil {
		return err
	}

	for _, seg := range segs {
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

// Split 将字符串解析成 Segment 数组
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
//  /posts/\{{id}/email ==> /posts/{, {id}/email
//  /posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func (i *Interceptors) Split(str string) ([]*Segment, error) {
	if str == "" {
		return nil, errors.New("参数 str 不能为空")
	}

	ss := splitString(str)
	segs := make([]*Segment, 0, len(ss))
	var lastFlag bool
	names := make(map[string]int, len(ss))

	for _, s := range ss {
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

		segs = append(segs, seg)
	}

	return segs, nil
}

func splitString(str string) []string {
	ss := make([]string, 0, strings.Count(str, string(startByte))+1)

	var end int
	for {
		start := strings.IndexByte(str[end:], startByte)
		if start == -1 {
			ss = append(ss, str)
			break
		} else if start > 0 {
			ss = append(ss, str[:start+end])
			str = str[start+end:]
		}

		end = strings.IndexByte(str, endByte)
		if end == -1 {
			ss = append(ss, str)
			break
		}
	}

	return ss
}
