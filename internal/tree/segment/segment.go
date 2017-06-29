// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package segment

import (
	"bytes"
	"fmt"
	"regexp"
	"regexp/syntax"
	"strings"
)

// Segment 表示路由中的分段内容。
type Segment interface {
	Type() Type
	Pattern() string
	Endpoint() bool
	Match(path string) (bool, string)
	Params(path string, params map[string]string) string
	URL(buf *bytes.Buffer, params map[string]string) error
}

type str struct {
	pattern string
}

type named struct {
	pattern string
	name    string
	suffix  string

	// 用于表示当前节点是否为终点，仅对 nodeType 为 TypeRegexp 和 TypeNamed 有用。
	// 此值为 true，该节点的优先级会比同类型的节点低，以便优先对比其它非最终节点。
	endpoint bool
}

type reg struct {
	pattern    string
	endpoint   bool
	expr       *regexp.Regexp
	syntaxExpr *syntax.Regexp
}

// New 将字符串转换为一个 Segment 实例。
// 调用者需要确保 str 语法正确。
func New(str string) (Segment, error) {
	typ := stringType(str)
	switch typ {
	case TypeNamed:
		return newNamed(str)
	case TypeString:
		return newStr(str)
	case TypeRegexp:
		return newReg(str)
	default:
		return nil, fmt.Errorf("无效的节点类型 %d", typ)
	}
}

func newStr(s string) (Segment, error) {
	return &str{
		pattern: s,
	}, nil
}

func (s *str) Type() Type {
	return TypeString
}

func (s *str) Pattern() string {
	return s.pattern
}

func (s *str) Endpoint() bool {
	return false
}

func (s *str) Match(path string) (bool, string) {
	if strings.HasPrefix(path, s.pattern) {
		return true, path[len(s.pattern):]
	}

	return false, path
}

func (s *str) Params(path string, params map[string]string) string {
	return path[len(s.pattern):]
}

func (s *str) URL(buf *bytes.Buffer, params map[string]string) error {
	buf.WriteString(s.pattern)
	return nil
}

func newNamed(str string) (Segment, error) {
	endIndex := strings.IndexByte(str, NameEnd)
	if endIndex == -1 {
		return nil, fmt.Errorf("无效的路由语法：%s", str)
	}

	return &named{
		pattern:  str,
		endpoint: str[len(str)-1] == NameEnd,
		name:     str[1:endIndex],
		suffix:   str[endIndex+1:],
	}, nil
}

func (n *named) Type() Type {
	return TypeNamed
}

func (n *named) Pattern() string {
	return n.pattern
}

func (n *named) Endpoint() bool {
	return n.endpoint
}

func (n *named) Match(path string) (bool, string) {
	if n.endpoint {
		return true, path[:0]
	}

	index := strings.Index(path, n.suffix)
	if index > 0 { // 为零说明前面没有命名参数，肯定不正确
		return true, path[index+len(n.suffix):]
	}

	return false, path
}

func (n *named) Params(path string, params map[string]string) string {
	if n.Endpoint() {
		params[n.name] = path
		return ""
	}

	index := strings.Index(path, n.suffix)
	params[n.name] = path[:index]
	return path[index+len(n.suffix):]
}

func (n *named) URL(buf *bytes.Buffer, params map[string]string) error {
	param, exists := params[n.name]
	if !exists {
		return fmt.Errorf("未找到参数 %s 的值", n.name)
	}
	buf.WriteString(param)
	buf.WriteString(n.suffix) // 如果是 endpoint suffix 肯定为空
	return nil
}

func newReg(str string) (Segment, error) {
	r := repl.Replace(str)
	expr, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}

	syntaxExpr, err := syntax.Parse(r, syntax.Perl)
	if err != nil {
		return nil, err
	}

	return &reg{
		pattern:    str,
		expr:       expr,
		syntaxExpr: syntaxExpr,
		endpoint:   str[len(str)-1] == NameEnd,
	}, nil
}

func (r *reg) Type() Type {
	return TypeRegexp
}

func (r *reg) Pattern() string {
	return r.pattern
}

func (r *reg) Endpoint() bool {
	return r.endpoint
}

func (r *reg) Match(path string) (bool, string) {
	loc := r.expr.FindStringIndex(path)
	if loc == nil || loc[0] != 0 { // 不匹配
		return false, path
	}

	if loc[1] == len(path) {
		return true, path[:0]
	}
	return true, path[loc[1]+1:]
}

func (r *reg) Params(path string, params map[string]string) string {
	subexps := r.expr.SubexpNames()
	args := r.expr.FindStringSubmatch(path)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			params[name] = args[index]
		}
	}

	return path[len(args[0]):]
}

func (r *reg) URL(buf *bytes.Buffer, params map[string]string) error {
	url := r.syntaxExpr.String()
	subs := append(r.syntaxExpr.Sub, r.syntaxExpr)
	for _, sub := range subs {
		if len(sub.Name) == 0 {
			continue
		}

		param, exists := params[sub.Name]
		if !exists {
			return fmt.Errorf("未找到参数 %v 的值", sub.Name)
		}
		url = strings.Replace(url, sub.String(), param, -1)
	}

	_, err := buf.WriteString(url)
	return err
}

// 获取字符串的类型。调用者需要确保 str 语法正确。
func stringType(str string) Type {
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
