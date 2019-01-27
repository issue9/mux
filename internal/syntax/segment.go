// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/issue9/mux/v2/params"
)

// 路由项字符串中的几个特殊字符定义
const (
	start     = '{'  // 命名或是正则参数的起始字符
	end       = '}'  // 命名或是正则参数的结束字符
	separator = ':'  // 正则参数中名称和正则的分隔符
	escape    = '\\' // 路由项中的转义字符
)

// Segment 路由项被拆分之后的值
type Segment struct {
	Value string
	Type  Type

	// 是否为最终结点
	//
	// 在命名路由项中，如果以 {path} 等结尾，则表示可以匹配任意剩余的字符。
	// 此值表示当前节点是否为此种类型。该类型的节点在匹配时，优先级可能会比较低。
	Endpoint bool

	// 当前节点的参数名称，比如 "{id}/author"，
	// 则此值为 "id"，非字符串节点有用。
	Name string

	// 保存参数名之后的字符串，比如 "{id}/author" 此值为 "/author"，
	// 仅对非字符串节点有效果，若 endpoint 为 false，则此值也不空。
	Suffix string

	// 正则表达式特有参数，用于缓存当前节点的正则编译结果。
	expr *regexp.Regexp
}

// NewSegment 声明新的 Segment 变量
func NewSegment(val string) *Segment {
	seg := &Segment{
		Value: val,
		Type:  getType(val),
	}

	switch seg.Type {
	case Named:
		index := strings.IndexByte(val, end)
		seg.Name = val[1:index]
		seg.Suffix = val[index+1:]
		seg.Endpoint = isEndpoint(val)
	case Regexp:
		seg.expr = regexp.MustCompile(repl.Replace(val))
		seg.Name = val[1:strings.IndexByte(val, separator)]
		seg.Suffix = val[strings.IndexByte(val, end)+1:]
	}

	return seg
}

// Similarity 与 s1 的相似度，-1 表示完全相同，0 表示完全不相同，
// 其它大于零的值，越大，表示相似度越高。
func (seg *Segment) Similarity(s1 *Segment) int {
	if s1.Value == seg.Value { // 有完全相同的节点
		return -1
	}

	return longestPrefix(s1.Value, seg.Value)
}

// Split 从 pos 位置拆分为两个
//
// pos 位置的字符归属于后一个元素。
func (seg *Segment) Split(pos int) []*Segment {
	return []*Segment{
		NewSegment(seg.Value[:pos]),
		NewSegment(seg.Value[pos:]),
	}
}

// Match 路径是否与当前节点匹配
//
// path 的起始部分如果与当前内容匹配，同返回 true 以及剩下未匹配的内容。
// 否则返回 false 以及 path 本身。
func (seg *Segment) Match(path string, params params.Params) (bool, string) {
	switch seg.Type {
	case String:
		if strings.HasPrefix(path, seg.Value) {
			return true, path[len(seg.Value):]
		}
	case Named:
		if seg.Endpoint {
			params[seg.Name] = path
			return true, path[:0]
		}

		// 为零说明前面没有命名参数，肯定不能与当前内容匹配
		if index := strings.Index(path, seg.Suffix); index > 0 {
			params[seg.Name] = path[:index]
			return true, path[index+len(seg.Suffix):]
		}
	case Regexp:
		locs := seg.expr.FindStringSubmatchIndex(path)
		if locs == nil || locs[0] != 0 { // 不匹配
			return false, path
		}

		params[seg.Name] = path[:locs[3]]
		return true, path[locs[1]:]
	}

	return false, path
}

// 获取两个字符串之间相同的前缀字符串的长度，
// 不会从 {} 中间被分开，正则表达式与之后的内容也不再分隔。
func longestPrefix(s1, s2 string) int {
	l := len(s1)
	if len(s2) < l {
		l = len(s2)
	}

	startIndex := -10
	endIndex := -10
	state := end
	for i := 0; i < l; i++ {
		switch s1[i] {
		case start:
			startIndex = i
			state = start
		case end:
			state = end
			endIndex = i
		}

		if s1[i] != s2[i] {
			if state != end { // 不从命名参数中间分隔
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

// Split 将字符串解析成字符串数组。
//
// 以 { 为分界线进行分割。比如
//  /posts/{id}/email ==> /posts/, {id}/email
//  /posts/\{{id}/email ==> /posts/{, {id}/email
//  /posts/{year}/{id}.html ==> /posts/, {year}/, {id}.html
func Split(str string) []*Segment {
	if str == "" {
		panic("参数 str 不能为空")
	}

	ss := make([]*Segment, 0, strings.Count(str, string(start))+1)

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case start:
			start := state.start
			state.setStart(i)

			if i == 0 { // 以 { 开头
				continue
			}

			ss = append(ss, NewSegment(str[start:i]))
		case separator:
			state.setSeparator(i)
		case end:
			state.setEnd(i)
		}

		if state.err != "" {
			panic(state.err)
		}
	} // end for

	if state.err != "" {
		panic(state.err)
	}

	if state.start < len(str) {
		if state.state != end {
			panic(fmt.Sprintf("缺少 %s 字符", string(end)))
		}

		ss = append(ss, NewSegment(str[state.start:]))
	}

	return ss
}

// IsWell 检测格式是否正确
func IsWell(str string) string {
	if str == "" {
		return "参数 str 不能为空"
	}

	state := newState()
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case start:
			state.setStart(i)

			if i == 0 { // 以 { 开头
				continue
			}
		case separator:
			state.setSeparator(i)
		case end:
			state.setEnd(i)
		}

		if state.err != "" {
			return state.err
		}
	} // end for

	if state.err != "" {
		return state.err
	}

	if state.start < len(str) {
		if state.state != end {
			return fmt.Sprintf("缺少 %s 字符", string(end))
		}
	}

	return ""
}
