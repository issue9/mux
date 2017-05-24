// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/issue9/mux/internal/entry"
	"github.com/issue9/mux/internal/method"
	"github.com/issue9/mux/internal/syntax"
)

const size = 10

// Byte 按字符进行分类的 entry.Entry 列表。
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
	entries map[byte]entries // TODO go1.9 改为 sync.Map
}

// NewByte 声明一个 Byte 实例
func NewByte(disableOptions bool) *Byte {
	return &Byte{
		disableOptions: disableOptions,
		entries:        make(map[byte]entries, 50),
	}
}

// Clean 清除所有的路由项，在 prefix 不为空的情况下，
// 则为删除所有路径前缀为 prefix 的匹配项。
func (b *Byte) Clean(prefix string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(prefix) == 0 {
		b.entries = make(map[byte]entries, 50)
		return
	}

	for _, es := range b.entries {
		es.clean(prefix)
	}
}

// Remove 移除指定的路由项。
//
// 当未指定 methods 时，将删除所有 method 匹配的项。
// 指定错误的 methods 值，将自动忽略该值。
func (b *Byte) Remove(pattern string, methods ...string) {
	if len(methods) == 0 {
		methods = method.Supported
	}
	index := b.byteIndex(pattern)

	b.mu.RLock()
	es, found := b.entries[index]
	b.mu.RUnlock()

	if !found {
		return
	}

	es.remove(pattern, methods...)

	if es.len() < size { // 数量少，改用 priority
		if p, ok := es.(*slash); ok {
			b.mu.Lock()
			b.entries[index] = p.toPriority()
			b.mu.Unlock()
		}
	}
}

// Add 添加一条路由数据。
//
// pattern 为路由匹配模式，可以是正则匹配也可以是字符串匹配；
// methods 为可以匹配的请求方法，默认为 method.Default 中的所有元素，
// 可以为 method.Supported 中的所有元素。
// 当 h 或是 pattern 为空时，将触发 panic。
func (b *Byte) Add(pattern string, h http.Handler, methods ...string) error {
	if len(pattern) == 0 {
		return errors.New("参数 pattern 不能为空")
	}
	if h == nil {
		return errors.New("参数 h 不能为空")
	}

	if byteCount('/', pattern) > maxSlashSize {
		return fmt.Errorf("最多只能有 %d 个 / 字符", maxSlashSize)
	}

	if len(methods) == 0 {
		methods = method.Default
	}

	s, err := syntax.New(pattern)
	if err != nil {
		return err
	}

	index := b.byteIndex(pattern)

	b.mu.Lock()
	defer b.mu.Unlock()
	es, found := b.entries[index]
	if !found {
		es = newPriority()
		b.entries[index] = es
	}

	if err = es.add(b.disableOptions, s, h, methods...); err != nil {
		return err
	}

	if es.len() > size { // 数量多，改用 slash
		if p, ok := es.(*priority); ok {
			es, err = p.toSlash()
			if err != nil {
				return err
			}
			b.entries[index] = es
		}
	}

	return nil
}

// Print 将内容以树状形式打印出来
func (b *Byte) Print() {
	for i, item := range b.entries {
		fmt.Println("#########", string(i))
		item.printDeep(0)
	}

	fmt.Println("+++++++++++++++++", len(b.entries))
}

// Entry 查找指定匹配模式下的 Entry，不存在，则声明新的
func (b *Byte) Entry(pattern string) (entry.Entry, error) {
	index := b.byteIndex(pattern)

	b.mu.RLock()
	defer b.mu.RUnlock()
	es, found := b.entries[index]
	if !found {
		es = newPriority()
		b.entries[index] = es
	}

	s, err := syntax.New(pattern)
	if err != nil {
		return nil, err
	}

	return es.entry(b.disableOptions, s)
}

// Match 查找与 path 最匹配的路由项以及对应的参数
func (b *Byte) Match(path string) (entry.Entry, map[string]string) {
	cnt := b.byteIndex(path)
	b.mu.RLock()
	es := b.entries[cnt]
	b.mu.RUnlock()
	if es != nil {
		if ety, ps := es.match(path); ety != nil {
			return ety, ps
		}
	}

	b.mu.RLock()
	es = b.entries[syntax.Start]
	b.mu.RUnlock()
	if es != nil {
		return es.match(path)
	}

	return nil, nil
}

// 计算 str 应该属于哪个 entries。
func (b *Byte) byteIndex(str string) byte {
	if len(str) < 2 {
		return 0
	}

	return str[1]
}
