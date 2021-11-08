// SPDX-License-Identifier: MIT

package syntax

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/issue9/mux/v5/params"
)

// 每次申请 params.Params 分配的大小
const defaultParamsCap = 5

// Segment 路由项被拆分之后的分段内容
type Segment struct {
	// 节点实际内容被拆分成以下几个部分，其组成方式如下：
	//  Value = {Name:Rule}Suffix
	// 其中 Name、Rule 和 Suffix 可能为空，但是 Name 和 Rule 不能同时为空。
	Value  string // 节点上的原始内容
	Name   string // 当前节点的参数名称，适用非字符节点。
	Rule   string // 节点的规则值，即 : 之后的部分，非字符串节点有效。
	Suffix string // 保存参数名之后的字符串，比如 "{id}/author" 此值为 "/author"，仅对非字符串节点有效果。

	// 节点类型
	Type Type

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

type MatchParam struct {
	Path   string
	Params params.Params
}

func (p *MatchParam) setParam(k, v string) {
	if p.Params == nil {
		p.Params = make(params.Params, defaultParamsCap)
	}
	p.Params[k] = v
}

// NewSegment 声明新的 Segment 变量
//
// 如果为非字符串类型的内容，应该是以 { 符号开头才是合法的。
func (i *Interceptors) NewSegment(val string) (*Segment, error) {
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
		return seg, nil
	}

	seg.Rule = val[separator+1 : end]
	if matcher, found := i.funcs[seg.Rule]; found {
		seg.Type = Interceptor
		seg.Name = val[start+1 : separator]
		seg.cleanName()
		seg.Suffix = val[end+1:]
		seg.Endpoint = val[len(val)-1] == endByte
		seg.matcher = matcher
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
	expr, err := regexp.Compile("(?" + name + seg.Rule + ")" + seg.Suffix)
	if err != nil {
		return nil, err
	}
	seg.expr = expr
	return seg, nil
}

func (seg *Segment) cleanName() {
	if seg.Name[0] == ignoreByte {
		seg.ignoreName = true
		seg.Name = seg.Name[1:]
	}
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

// Match 路径是否与当前节点匹配
//
// 如果正确匹配，则将剩余的未匹配字符串写入到 p.Path 并返回 true。
func (seg *Segment) Match(p *MatchParam) bool {
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
					p.setParam(seg.Name, p.Path)
				}
				p.Path = p.Path[:0]
				return true
			}
		} else if index := strings.Index(p.Path, seg.Suffix); index >= 0 {
			for {
				if val := p.Path[:index]; seg.matcher(val) {
					if !seg.ignoreName {
						p.setParam(seg.Name, val)
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
		if loc := seg.expr.FindStringSubmatchIndex(p.Path); loc != nil && loc[0] == 0 {
			if !seg.ignoreName {
				p.setParam(seg.Name, p.Path[:loc[3]])
			}
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
