// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"container/list"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// entry列表。
type entries struct {
	list  *list.List        // 路由列表，静态路由在前，正则路由在后。
	named map[string]*entry // 路由的命名列表，方便查找。
}

type entry struct {
	pattern   string         // 匹配字符串
	expr      *regexp.Regexp // 若是正则匹配，则会被pattern字段转换成正则表达式，并保存在此变量中
	hasParams bool           // 是否拥有命名路由参数，仅在expr不为nil的时候有用
	handler   http.Handler
}

func newEntries() *entries {
	return &entries{
		list:  list.New(),
		named: map[string]*entry{},
	}
}

// 向entry列表添加一个路由项。
func (es *entries) add(pattern string, h http.Handler) error {
	if _, found := es.named[pattern]; found {
		return errors.New("该模式的路由项已经存在")
	}

	e := newEntry(pattern, h)
	es.named[pattern] = e

	if e.expr == nil { // 静态路由，在前端插入
		es.list.PushFront(e)
	} else { // 正则路由，在后端插入
		es.list.PushBack(e)
	}

	return nil
}

// 从entry列表中移除一个路由项。
// 若没有与pattern匹配的路由项，将不发生任何操作。
func (es *entries) remove(pattern string) {
	if _, found := es.named[pattern]; !found {
		return
	}

	delete(es.named, pattern)

	for item := es.list.Front(); item != nil; item = item.Next() {
		e := item.Value.(*entry)
		if e.pattern == pattern {
			es.list.Remove(item)
			return
		}
	}
}

// 匹配程度
//  -1 表示完全不匹配
//  0  表示完全匹配
//  >0 表示部分匹配，值越小表示匹配程度越高。
func (e *entry) match(url string) int {
	if e.expr == nil { // 静态匹配
		if len(url) < len(e.pattern) {
			return -1
		}

		if e.pattern == url[:len(e.pattern)] {
			return len(url) - len(e.pattern)
		}

		return -1
	}

	// 正则匹配
	if loc := e.expr.FindStringIndex(url); loc != nil {
		return len(url) - loc[1]
	}
	return -1
}

// 将url与当前的表达式进行匹配，返回其命名路由参数的值。若不匹配，则返回nil
func (e *entry) getParams(url string) map[string]string {
	if !e.hasParams {
		return nil
	}

	// 正确匹配正则表达式，则获相关的正则表达式命名变量。
	mapped := make(map[string]string)
	subexps := e.expr.SubexpNames()
	args := e.expr.FindStringSubmatch(url)
	for index, name := range subexps {
		if len(name) > 0 && index < len(args) {
			mapped[name] = args[index]
		}
	}
	return mapped
}

// 声明一个entry实例
// pattern 匹配内容。
// h 对应的http.Handler，外层调用者确保该值不能为nil.
func newEntry(pattern string, h http.Handler) *entry {
	strs := split(pattern)

	if len(strs) == 1 { // 静态路由
		return &entry{
			pattern: pattern,
			handler: h,
		}
	}

	pattern, hasParams := toPattern(strs)
	return &entry{
		pattern:   pattern,
		expr:      regexp.MustCompile(pattern),
		hasParams: hasParams,
		handler:   h,
	}
}

// 将strs按照顺序合并成一个正则表达式
// 返回参数正则表达式的字符串和一个bool值用以表式正则中是否包含了命名匹配。
func toPattern(strs []string) (string, bool) {
	pattern := ""
	hasParams := false

	for _, v := range strs {
		lastIndex := len(v) - 1
		if v[0] != '{' || v[lastIndex] != '}' { // 普通字符串
			pattern += v
			continue
		}

		v = v[1:lastIndex] // 去掉首尾的{}符号

		index := strings.IndexByte(v, ':')
		if index < 0 { // 只存在命名，而不存在正则表达式，默认匹配[^/]
			pattern += "(?P<" + v + ">[^/]+)"
			hasParams = true
			continue
		}

		if index == 0 { // 不存在命名，但有正则表达式
			pattern += v[1:]
			continue
		}

		pattern += "(?P<" + v[:index] + ">" + v[index+1:] + ")"
		hasParams = true
	}

	return pattern, hasParams
}

// 将str以{和}为分隔符进行分隔。
// 符号{和}必须成对出现，且不能嵌套，否则结果是未知的。
func split(str string) []string {
	ret := []string{}
	var seq byte = '{'

	for {
		if len(str) == 0 { // 没有更多字符了，结束
			break
		}

		index := strings.IndexByte(str, seq)
		if index < 0 { // 未找到分隔符，结束
			ret = append(ret, str)
			break
		}

		if seq == '}' { // 将}字符留在当前字符串中
			index++
		}

		if index > 0 { // 为零表示当前字符串为空，无须理会。
			ret = append(ret, str[:index])
			str = str[index:]
		}

		if seq == '{' {
			seq = '}'
		} else {
			seq = '{'
		}
	}

	return ret
}
