// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"regexp"
	"strings"
)

type entryer interface {
	match(url string) (int, map[string]string)
}

// /post/1
type staticEntry string

func (e staticEntry) match(url string) (int, map[string]string) {
	if len(url) < len(e) {
		return -1, nil
	}

	if string(e) == url[:len(e)] {
		return len(url) - len(e), nil
	}

	return -1, nil
}

type node struct {
	ntype int // 1:普通字符串;2:匹配任意值;3:正则表达式
	expr  *regexp.Regexp
	value string
}

func (n node) isMatch(str string) bool {
	switch n.ntype {
	case 1:
		return n.value == str
	case 2:
		return n.value == str
	case 3:
		return n.expr.MatchString(str)
	default:
		panic("node.isMatch:错误的ntype")
	}
}

// /blog/{post:\\w+}-{:\\d+}
// /blog/(<?P<post>\\w+)-(\\d+)
type regexpEntry []node

// {action:\\w+}{id:\\d+}/page
// post1/page
func (e regexpEntry) match(url string) (int, map[string]string) {
	for _, v := range e {
		if v.ntype == 1 {
			if len(v.value) > len(url) || v.value != url[:len(v.value)] {
				return -1, nil
			}
			url = url[len(v.value):]
			continue
		}

		if v.ntype == 2 {
		}
	}

	return 0, nil
}

// 将pattern以{和}为分隔符进行分隔。
// 符号{和}必须成对出现，且不能嵌套，否则结果是未知的。
func split(str string) []string {
	ret := []string{}
	var seq byte = '{'

	for {
		index := strings.IndexByte(str, seq)
		if len(str) == 0 {
			break
		}

		if index < 0 {
			ret = append(ret, str)
			break
		}

		if seq == '}' {
			index++
		}
		if index > 0 {
			ret = append(ret, str[:index])
		}
		str = str[index:]

		if seq == '{' {
			seq = '}'
		} else {
			seq = '{'
		}
	}

	return ret
}

func newEntry(pattern string) entryer {
	strs := split(pattern)

	if len(strs) == 1 { // 静态路由
		return staticEntry(pattern)
	}

	nodes := make([]node, 0, len(strs))
	for _, v := range strs {
		lastIndex := len(v) - 1
		if v[0] != '{' || v[lastIndex] != '}' { // 普通字符串
			nodes = append(nodes, node{ntype: 1, value: v, expr: nil})
			continue
		}

		v = v[1:lastIndex] // 去掉首尾的{}符号

		index := strings.IndexByte(v, ':')
		if index < 0 { // 不存在正则表达式
			nodes = append(nodes, node{ntype: 2, value: v, expr: nil})
			continue
		}

		expr := regexp.MustCompile(v[index+1:])
		nodes = append(nodes, node{ntype: 3, value: v[:index], expr: expr})
	}
	return &regexpEntry{}
}
