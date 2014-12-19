// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"sync"
)

// 与http.Request相关联的上下文环境
type context struct {
	sync.Mutex
	items map[interface{}]interface{}
}

// 查找key对应的值。在没有查到的情况下，found返回false
func (ctx *context) Get(key interface{}) (val interface{}, found bool) {
	ctx.Lock()
	defer ctx.Unlock()

	val, found = ctx.items[key]
	return
}

// 功能与Get()相同，在没有找到相关值的情况下，会返回def，但该真不会写
// 入到context中，下次用Get()依然会返回false
func (ctx *context) MustGet(key, def interface{}) interface{} {
	ctx.Lock()
	defer ctx.Unlock()

	val, found := ctx.items[key]
	if !found {
		return def
	} else {
		return val
	}
}

// 设置或是添加一个键值对。
func (ctx *context) Set(key, val interface{}) {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.items[key] = val
}

// 添加一个键值对，若该键名已经存在，则作任何操作，
// 并且ok返回false，表示操作没有成功
func (ctx *context) Add(key, val interface{}) (ok bool) {
	ctx.Lock()
	defer ctx.Unlock()

	_, found := ctx.items[key]
	if found {
		return false
	}

	ctx.items[key] = val
	return true
}

// TODO(caixw) context对象池，对象太小，性能并没有提升，是否需要使用该特性？
var ctxFree = sync.Pool{
	New: func() interface{} { return &context{items: make(map[interface{}]interface{})} },
}

var contexts = &contextMap{
	items: make(map[*http.Request]*context),
}

type contextMap struct {
	sync.Mutex
	items map[*http.Request]*context
}

// 获取或是新建context
func GetContext(r *http.Request) *context {
	contexts.Lock()
	defer contexts.Unlock()

	ctx, found := contexts.items[r]
	if found {
		return ctx
	}

	ctx = ctxFree.Get().(*context)
	if len(ctx.items) > 0 { // 确保新项目的items为0
		ctx.items = make(map[interface{}]interface{})
	}
	contexts.items[r] = ctx
	return ctx
}

// 释放context
func freeContext(r *http.Request) {
	contexts.Lock()
	defer contexts.Unlock()

	ctx, found := contexts.items[r]
	if !found {
		return
	}

	delete(contexts.items, r)
	ctxFree.Put(ctx)
}
