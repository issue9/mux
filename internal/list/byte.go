// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"errors"
	"net/http"
	"sync"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
)

// Byte entry.Entry 列表。
type Byte struct {
	disableOptions bool
	mu             sync.RWMutex

	// entries 是按路由项首字母进行第一次分类，
	// 这样在进行路由匹配时，可以减少大量的时间：
	//  /posts/{id}              // p
	//  /tags/{name}             // t
	//  /posts/{id}/author       // p
	//  /posts/{id}/author/*     // p
	// 比如以上路由项，如果要查找 /posts/1 只需要比较 p
	// 中的数据就行，如果需要匹配 /tags/abc.html 则只需要比较 t。
	entries map[byte]*slash // TODO go1.9 改为 sync.Map
}

// NewByte 声明一个 Byte 实例
func NewByte(disableOptions bool) *Byte {
	return &Byte{
		disableOptions: disableOptions,
		entries:        make(map[byte]*slash, 28), // 26 + '{' + 0
	}
}

// Clean 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (l *Byte) Clean(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(prefix) == 0 {
		l.entries = make(map[byte]*slash, 28)
		return
	}

	for _, es := range l.entries {
		es.Clean(prefix)
	}
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (l *Byte) Remove(pattern string, methods ...string) {
	if len(methods) == 0 {
		methods = method.Supported
	}

	l.mu.RLock()
	es, found := l.entries[l.slashIndex(pattern)]
	l.mu.RUnlock()

	if !found {
		return
	}

	es.Remove(pattern, methods...)
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配；
// methods 为可以匹配的请求方法，默认为 method.Default 中的所有元素，
// 可以为 method.Supported 中的所有元素。
// 当 h 或是 pattern 为空时，将触发 panic。
func (l *Byte) Add(pattern string, h http.Handler, methods ...string) error {
	if len(pattern) == 0 {
		return errors.New("参数 pattern 不能为空")
	}
	if h == nil {
		return errors.New("参数 h 不能为空")
	}

	if len(methods) == 0 {
		methods = method.Default
	}

	index := l.slashIndex(pattern)

	l.mu.Lock()
	defer l.mu.Unlock()
	es, found := l.entries[index]
	if !found {
		es = newSlash(l.disableOptions)
		l.entries[index] = es
	}

	return es.Add(pattern, h, methods...)
}

// Entry 查找指定匹配模式下的 Entry，不存在，则声明新的
func (l *Byte) Entry(pattern string) (entry.Entry, error) {
	index := l.slashIndex(pattern)

	l.mu.RLock()
	defer l.mu.RUnlock()
	es, found := l.entries[index]
	if !found {
		es = newSlash(l.disableOptions)
		l.entries[index] = es
	}

	return es.Entry(pattern)
}

// Match 查找与 path 最匹配的路由项以及对应的参数
func (l *Byte) Match(path string) (entry.Entry, map[string]string) {
	cnt := l.slashIndex(path)
	l.mu.RLock()
	es := l.entries[cnt]
	l.mu.RUnlock()
	if es != nil {
		if ety, ps := es.Match(path); ety != nil {
			return ety, ps
		}
	}

	l.mu.RLock()
	es = l.entries['{']
	l.mu.RUnlock()
	if es != nil {
		return es.Match(path)
	}

	return nil, nil
}

// Len
func (l *Byte) Len() int {
	ret := 0
	for _, item := range l.entries {
		ret += item.len()
	}
	return ret
}

// 计算 str 应该属于哪个 entries。
func (l *Byte) slashIndex(str string) byte {
	if len(str) < 2 {
		return 0
	}

	return str[1]
}
