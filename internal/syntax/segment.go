// SPDX-License-Identifier: MIT

package syntax

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/issue9/mux/v4/interceptor"
	"github.com/issue9/mux/v4/params"
)

// Segment 路由项被拆分之后的分段内容
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

	// 正则表达式的拦截器的处理函数
	matcher interceptor.MatchFunc
}

// NewSegment 声明新的 Segment 变量
//
// 如果为非字符串类型的内容，应该是以 { 符号开头才是合法的。
func NewSegment(val string) (*Segment, error) {
	seg := &Segment{
		Value: val,
		Type:  String,
	}

	start := strings.IndexByte(val, startByte)
	end := strings.IndexByte(val, endByte)
	if start == -1 || end == -1 {
		return seg, nil
	}

	if start > end || start+1 == end { // }{ 或是  {}
		return nil, fmt.Errorf("无效的语法：%s", val)
	}

	separator := strings.IndexByte(val, separatorByte)
	if separator == -1 || separator+1 == end || separator > end { // {abc} 或是 {abc:} 或是 {abc}:
		seg.Name = val[start+1 : end]
		if separator != -1 && separator < end {
			seg.Name = val[start+1 : separator]
		}

		seg.Type = Named
		seg.Suffix = val[end+1:]
		seg.Endpoint = val[len(val)-1] == endByte
		seg.matcher = func(string) bool { return true }
		return seg, nil
	}

	matcher, found := interceptor.Get(val[separator+1 : end])
	if found {
		seg.Type = Interceptor
		seg.Name = val[start+1 : separator]
		seg.Suffix = val[end+1:]
		seg.Endpoint = val[len(val)-1] == endByte
		seg.matcher = matcher
		return seg, nil
	}

	seg.Type = Regexp
	seg.Name = val[start+1 : separator]
	seg.Suffix = val[end+1:]
	expr, err := regexp.Compile("(?P<" + seg.Name + ">" + val[separator+1:end] + ")" + seg.Suffix)
	if err != nil {
		return nil, err
	}
	seg.expr = expr
	return seg, nil
}

// Similarity 与 s1 的相似度，-1 表示完全相同，
// 其它大于等于零的值，越大，表示相似度越高。
func (seg *Segment) Similarity(s1 *Segment) int {
	if s1.Value == seg.Value { // 有完全相同的节点
		return -1
	}

	return longestPrefix(s1.Value, seg.Value)
}

// Split 从 pos 位置拆分为两个
//
// pos 位置的字符归属于后一个元素。
func (seg *Segment) Split(pos int) ([]*Segment, error) {
	s1, err := NewSegment(seg.Value[:pos])
	if err != nil {
		return nil, err
	}
	s2, err := NewSegment(seg.Value[pos:])
	if err != nil {
		return nil, err
	}
	return []*Segment{s1, s2}, nil
}

// Match 路径是否与当前节点匹配
//
// 如果正确匹配，则返回 path 剩余部分的起始位置。
// params 表示匹配完成之后，从地址中获取的参数值。
func (seg *Segment) Match(path string, params params.Params) int {
	switch seg.Type {
	case String:
		if strings.HasPrefix(path, seg.Value) {
			return len(seg.Value)
		}
	case Interceptor, Named:
		if seg.Endpoint {
			if seg.matcher(path) {
				params[seg.Name] = path
				return len(path)
			}
		} else if index := strings.Index(path, seg.Suffix); index >= 0 {
			for {
				if val := path[:index]; seg.matcher(val) {
					params[seg.Name] = val
					return index + len(seg.Suffix)
				}

				i := strings.Index(path[index+len(seg.Suffix):], seg.Suffix)
				if i < 0 {
					return -1
				}
				index += i + len(seg.Suffix)
			}
		}
	case Regexp:
		if loc := seg.expr.FindStringSubmatchIndex(path); loc != nil && loc[0] == 0 {
			params[seg.Name] = path[:loc[3]]
			return loc[1]
		}
	}

	return -1
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
	state := endByte
	for i := 0; i < l; i++ {
		switch s1[i] {
		case startByte:
			startIndex = i
			state = startByte
		case endByte:
			state = endByte
			endIndex = i
		}

		if s1[i] != s2[i] {
			if state != endByte || // 不从命名参数中间分隔
				endIndex == i { // 命名参数之后必须要有一个或以上的普通字符
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
