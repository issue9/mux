// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package list 提供了对 entry.Entry 元素的存储、匹配等功能。
package list

import (
	"net/http"
	"sync"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
	"github.com/issue9/mux/internal/syntax"
)

const wildcardIndex = -1

// slash entry.Entry 列表。
type slash struct {
	disableOptions bool
	mu             sync.RWMutex

	// entries 是按路由项中 / 字符的数量进行分类，
	// 这样在进行路由匹配时，可以减少大量的时间：
	//  /posts/{id}              // 2
	//  /tags/{name}             // 2
	//  /posts/{id}/author       // 3
	//  /posts/{id}/author/*     // -1
	// 比如以上路由项，如果要查找 /posts/1 只需要比较 2
	// 中的数据就行，如果需要匹配 /tags/abc/1.html 则只需要比较 3。
	entries map[int]*entries // TODO go1.9 改为 sync.Map
}

// New 声明一个 slash 实例
func newSlash(disableOptions bool) *slash {
	return &slash{
		disableOptions: disableOptions,
		entries:        make(map[int]*entries, 20),
	}
}

// Clean 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (l *slash) Clean(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(prefix) == 0 {
		l.entries = make(map[int]*entries, 20)
		return
	}

	for _, es := range l.entries {
		es.clean(prefix)
	}
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (l *slash) Remove(pattern string, methods ...string) {
	s, err := syntax.New(pattern)
	if err != nil { // 错误的语法，肯定不存在于现有路由项，可以直接返回
		return
	}

	if len(methods) == 0 {
		methods = method.Supported
	}

	l.mu.RLock()
	es, found := l.entries[l.entriesIndex(s)]
	l.mu.RUnlock()

	if !found {
		return
	}

	es.remove(pattern, methods...)
}

// 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配；
// methods 为可以匹配的请求方法，默认为 method.Default 中的所有元素，
// 可以为 method.Supported 中的所有元素。
func (l *slash) add(s *syntax.Syntax, h http.Handler, methods ...string) error {
	index := l.entriesIndex(s)

	l.mu.Lock()
	defer l.mu.Unlock()
	es, found := l.entries[index]
	if !found {
		es = newEntries(l.disableOptions)
		l.entries[index] = es
	}

	return es.add(s, h, methods...)
}

// 查找指定匹配模式下的 entry.Entry，不存在，则声明新的
func (l *slash) entry(s *syntax.Syntax) (entry.Entry, error) {
	index := l.entriesIndex(s)

	l.mu.RLock()
	defer l.mu.RUnlock()
	es, found := l.entries[index]
	if !found {
		es = newEntries(l.disableOptions)
		l.entries[index] = es
	}

	return es.entry(s)
}

// 查找与 path 最匹配的路由项以及对应的参数
func (l *slash) match(path string) (entry.Entry, map[string]string) {
	cnt := byteCount('/', path)
	l.mu.RLock()
	es := l.entries[cnt]
	l.mu.RUnlock()
	if es != nil {
		if ety, ps := es.match(path); ety != nil {
			return ety, ps
		}
	}

	l.mu.RLock()
	es = l.entries[wildcardIndex]
	l.mu.RUnlock()
	if es != nil {
		return es.match(path)
	}

	return nil, nil
}

// 计算 str 应该属于哪个 entries。
func (l *slash) entriesIndex(s *syntax.Syntax) int {
	if s.Wildcard || s.Type == syntax.TypeRegexp {
		return wildcardIndex
	}

	return byteCount('/', s.Pattern)
}

// 统计字符串包含的指定字符的数量
func byteCount(b byte, str string) int {
	ret := 0
	for i := 0; i < len(str); i++ {
		if str[i] == b {
			ret++
		}
	}

	return ret
}
