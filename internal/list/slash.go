// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"fmt"
	"net/http"
	"sync"

	"strings"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
	"github.com/issue9/mux/internal/syntax"
)

const wildcardIndex = -1

// 按 / 进行分类的 entry.Entry 列表
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
	entries map[int]*priority // TODO go1.9 改为 sync.Map
}

func newSlash(disableOptions bool) *slash {
	return &slash{
		disableOptions: disableOptions,
		entries:        make(map[int]*priority, 20),
	}
}

// entries.clean
func (l *slash) clean(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(prefix) == 0 {
		l.entries = make(map[int]*priority, 20)
		return
	}

	for _, es := range l.entries {
		es.clean(prefix)
	}
}

// entries.remove
func (l *slash) remove(pattern string, methods ...string) {
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

// entries.add
func (l *slash) add(s *syntax.Syntax, h http.Handler, methods ...string) error {
	index := l.entriesIndex(s)

	l.mu.Lock()
	defer l.mu.Unlock()
	es, found := l.entries[index]
	if !found {
		es = newPriority(l.disableOptions)
		l.entries[index] = es
	}

	return es.add(s, h, methods...)
}

// entries.entry
func (l *slash) entry(s *syntax.Syntax) (entry.Entry, error) {
	index := l.entriesIndex(s)

	l.mu.RLock()
	defer l.mu.RUnlock()
	es, found := l.entries[index]
	if !found {
		es = newPriority(l.disableOptions)
		l.entries[index] = es
	}

	return es.entry(s)
}

// entries.match
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

// entries.len
func (l *slash) len() int {
	ret := 0
	for _, es := range l.entries {
		ret += es.len()
	}

	return ret
}

func (l *slash) toPriority() entries {
	es := newPriority(l.disableOptions)
	for _, item := range l.entries {
		for _, i := range item.entries {
			es.entries = append(es.entries, i)
		}
	}

	return es
}

func (l *slash) addEntry(ety entry.Entry) error {
	s, err := syntax.New(ety.Pattern())
	if err != nil {
		return err
	}

	index := l.entriesIndex(s)

	l.mu.Lock()
	defer l.mu.Unlock()
	es, found := l.entries[index]
	if !found {
		es = newPriority(l.disableOptions)
		l.entries[index] = es
	}

	es.mu.Lock()
	es.entries = append(es.entries, ety)
	es.mu.Unlock()
	return nil
}

func (l *slash) printDeep(deep int) {
	fmt.Println(strings.Repeat(" ", deep*4), "---------slash")
	for _, item := range l.entries {
		item.printDeep(deep + 1)
	}
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
