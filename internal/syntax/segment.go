// SPDX-License-Identifier: MIT

package syntax

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Segment 路由项被拆分之后的分段内容
type Segment struct {
	// 节点实际内容被拆分成以下几个部分，其组成方式如下：
	//  Value = {Name:rule}Suffix
	// 其中 Name、rule 和 Suffix 可能为空，但是 Name 和 rule 不能同时为空。
	Value  string // 节点上的原始内容
	Name   string // 当前节点的参数名称，适用非字符节点。
	rule   string // 节点的规则值，即 : 之后的部分，非字符串节点有效。
	Suffix string // 保存参数名之后的字符串，比如 "{id}/author" 此值为 "/author"，非 endpoint 的拦截器和命名节点不能为空。。

	// TODO: Type、ambiguousLength、Endpoint 和 ignoreName 采用一个 int 字段代替

	// 节点类型
	Type Type

	// 保存着当前节点 Value 中除去名称之外的节点长度。
	// 当判断两个节点是否存在歧义时，会用到此值。
	ambiguousLength int16

	// 是否为最终结点
	//
	// 在命名和拦截类型的路由项中，如果以 {path} 等结尾，则表示可以匹配任意剩余的字符。
	// 此值表示当前节点是否为此种类型。该类型的节点在匹配时，优先级可能会比较低。
	Endpoint bool

	// 忽略名称
	ignoreName bool

	// 正则表达式特有参数，用于缓存当前节点的正则编译结果。
	expr *regexp.Regexp

	// 拦截器的处理函数
	matcher InterceptorFunc
}

// NewSegment 声明新的 Segment 变量
//
// 如果为非字符串类型的内容，应该是以 { 符号开头才是合法的。
func (i *Interceptors) NewSegment(val string) (*Segment, error) {
	if len(val) > math.MaxInt16 {
		return nil, fmt.Errorf("单个节点的长度不能超过 %d", math.MaxInt16)
	}

	seg := &Segment{
		Value: val,
		Type:  String,
	}

	start := strings.IndexByte(val, startByte)
	end := strings.IndexByte(val, endByte)
	if start == -1 || end == -1 {
		return seg, nil
	}

	separator := strings.IndexByte(val, separatorByte)
	if start > end || start+1 == end || // }{ 或是  {}
		(separator > 0 && start+1 == separator) { // {:rule}
		return nil, fmt.Errorf("无效的语法：%s", val)
	}

	if separator == -1 || separator+1 == end || separator > end { // {name} 或是 {name:} 或是 {name}:
		seg.Name = val[start+1 : end]
		if separator != -1 && separator < end {
			seg.Name = val[start+1 : separator]
		}

		seg.Type = Named
		seg.Suffix = val[end+1:]
		seg.Endpoint = val[len(val)-1] == endByte
		seg.matcher = func(string) bool { return true }
		seg.cleanName()
		seg.calcAmbiguousLength()
		return seg, nil
	}

	seg.rule = val[separator+1 : end]
	if matcher, found := i.funcs[seg.rule]; found {
		seg.Type = Interceptor
		seg.Name = val[start+1 : separator]
		seg.cleanName()
		seg.Suffix = val[end+1:]
		seg.Endpoint = val[len(val)-1] == endByte
		seg.matcher = matcher
		seg.calcAmbiguousLength()
		return seg, nil
	}

	seg.Type = Regexp
	seg.Name = val[start+1 : separator]
	seg.cleanName()
	seg.Suffix = val[end+1:]
	name := ":"
	if !seg.ignoreName {
		name = "P<" + seg.Name + ">"
	}
	expr, err := regexp.Compile("(?" + name + seg.rule + ")" + seg.Suffix)
	if err != nil {
		return nil, err
	}
	seg.expr = expr
	seg.calcAmbiguousLength()
	return seg, nil
}

func (seg *Segment) cleanName() {
	if seg.Name[0] == ignoreByte {
		seg.ignoreName = true
		seg.Name = seg.Name[1:]
	}
}

func (seg *Segment) calcAmbiguousLength() {
	seg.ambiguousLength = 2 // {}

	if seg.ignoreName { // -
		seg.ambiguousLength++
	}

	if seg.rule != "" {
		seg.ambiguousLength += int16(len(seg.rule))
		seg.ambiguousLength++ // 表示 :
	}

	if seg.Suffix != "" {
		seg.ambiguousLength += int16(len(seg.Suffix))
	}
}

// IsAmbiguous 判断两个节点是否存在歧义
//
// 即除了名称，其它都相同，如果两个节点符合此条件，
// 在相同路径下是无法判断到底要选哪一条路径的，应该避免此类节节点出现在同一路径上。
func (seg *Segment) IsAmbiguous(s2 *Segment) bool {
	if seg.ignoreName != s2.ignoreName {
		return seg.Endpoint == s2.Endpoint && seg.Type == s2.Type && seg.rule == s2.rule && seg.Suffix == s2.Suffix
	}
	return seg.Name != s2.Name &&
		seg.ambiguousLength == s2.ambiguousLength &&
		(seg.Endpoint == s2.Endpoint && seg.Type == s2.Type && seg.rule == s2.rule && seg.Suffix == s2.Suffix)
}

func (seg *Segment) AmbiguousLen() int16 {
	return seg.ambiguousLength + int16(len(seg.Name))
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
func (seg *Segment) Split(i *Interceptors, pos int) ([]*Segment, error) {
	s1, err := i.NewSegment(seg.Value[:pos])
	if err != nil {
		return nil, err
	}
	s2, err := i.NewSegment(seg.Value[pos:])
	if err != nil {
		return nil, err
	}
	return []*Segment{s1, s2}, nil
}

// Valid 验证 pattern 是否与当前节点匹配
func (seg *Segment) Valid(pattern string) bool {
	switch seg.Type {
	case Interceptor:
		return seg.matcher(pattern)
	case Regexp:
		pattern += seg.Suffix
		locs := seg.expr.FindStringIndex(pattern)
		return locs != nil && locs[1] == len(pattern)
	}
	return true
}

// Match 路径是否与当前节点匹配
//
// 如果正确匹配，则将剩余的未匹配字符串写入到 p.Path 并返回 true。
func (seg *Segment) Match(p *Params) bool {
	switch seg.Type {
	case String:
		if strings.HasPrefix(p.Path, seg.Value) {
			p.Path = p.Path[len(seg.Value):]
			return true
		}
	case Interceptor, Named:
		if seg.Endpoint {
			if seg.matcher(p.Path) {
				if !seg.ignoreName {
					p.Set(seg.Name, p.Path)
				}
				p.Path = p.Path[:0]
				return true
			}
		} else if index := strings.Index(p.Path, seg.Suffix); index >= 0 {
			for {
				if val := p.Path[:index]; seg.matcher(val) {
					if !seg.ignoreName {
						p.Set(seg.Name, val)
					}
					p.Path = p.Path[index+len(seg.Suffix):]
					return true
				}

				i := strings.Index(p.Path[index+len(seg.Suffix):], seg.Suffix)
				if i < 0 {
					return false
				}
				index += i + len(seg.Suffix)
			}
		}
	case Regexp:
		if seg.ignoreName {
			if loc := seg.expr.FindStringIndex(p.Path); loc != nil && loc[0] == 0 {
				p.Path = p.Path[loc[1]:]
				return true
			}
		} else if loc := seg.expr.FindStringSubmatchIndex(p.Path); loc != nil && loc[0] == 0 {
			p.Set(seg.Name, p.Path[:loc[3]]) // 只有 ignoreName == false，才会有捕获的值
			p.Path = p.Path[loc[1]:]
			return true
		}
	}

	return false
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
				endIndex+1 == i { // 命名参数之后必须要有一个或以上的普通字符
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
