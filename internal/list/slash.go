// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"fmt"
	"net/http"

	"strings"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
	"github.com/issue9/mux/internal/syntax"
)

const (
	maxSlashSize   = 20
	lastSlashIndex = maxSlashSize - 1
)

// 按 / 进行分类的 entry.Entry 列表
type slash struct {
	// entries 是按路由项中 / 字符的数量进行分类，索引值表示数量，
	// 元素表示对应的 priority 对象。这样在进行路由匹配时，可以减少大量的时间：
	//  /posts/{id}              // 2
	//  /tags/{name}             // 2
	//  /posts/{id}/author       // 3
	//  /posts/{id}/author/*     // -1
	// 比如以上路由项，如果要查找 /posts/1 只需要比较 2
	// 中的数据就行，如果需要匹配 /tags/abc/1.html 则只需要比较 3。
	entries []*priority
}

func newSlash() *slash {
	return &slash{
		entries: make([]*priority, maxSlashSize),
	}
}

// entries.clean
func (l *slash) clean(prefix string) {
	if len(prefix) == 0 {
		for i := 0; i < maxSlashSize; i++ {
			l.entries[i] = nil
		}
		return
	}

	for _, es := range l.entries {
		if es == nil {
			continue
		}
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

	es := l.entries[l.slashIndex(s)]
	if es == nil {
		return
	}

	es.remove(pattern, methods...)
}

// entries.add
func (l *slash) add(disableOptions bool, s *syntax.Syntax, h http.Handler, methods ...string) error {
	index := l.slashIndex(s)

	es := l.entries[index]
	if es == nil {
		es = newPriority()
		l.entries[index] = es
	}

	return es.add(disableOptions, s, h, methods...)
}

// entries.entry
func (l *slash) entry(disableOptions bool, s *syntax.Syntax) (entry.Entry, error) {
	index := l.slashIndex(s)

	es := l.entries[index]
	if es == nil {
		es = newPriority()
		l.entries[index] = es
	}

	return es.entry(disableOptions, s)
}

// entries.match
func (l *slash) match(path string) (entry.Entry, map[string]string) {
	es := l.entries[byteCount('/', path)]
	if es != nil {
		if ety, ps := es.match(path); ety != nil {
			return ety, ps
		}
	}

	es = l.entries[lastSlashIndex]
	if es != nil {
		return es.match(path)
	}

	return nil, nil
}

// 计算 str 应该属于哪个 entries。
func (l *slash) slashIndex(s *syntax.Syntax) int {
	if s.Wildcard || s.Type == syntax.TypeRegexp {
		return lastSlashIndex
	}

	return byteCount('/', s.Pattern)
}

// entries.len
func (l *slash) len() int {
	ret := 0
	for _, es := range l.entries {
		if es == nil {
			continue
		}
		ret += es.len()
	}

	return ret
}

func (l *slash) toPriority() entries {
	es := newPriority()
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

	index := l.slashIndex(s)

	es := l.entries[index]
	if es == nil {
		es = newPriority()
		l.entries[index] = es
	}

	es.entries = append(es.entries, ety)
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
