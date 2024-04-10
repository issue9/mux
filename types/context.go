// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"strconv"
	"sync"
)

var contextPool = &sync.Pool{New: func() any { return &Context{} }}

// Context 保存着路由匹配过程中的上下文关系
//
// Context 同时实现了 [Route] 接口。
type Context struct {
	Path        string // 实际请求的路径信息
	keys        []string
	vals        []string
	paramsCount int
	routerName  string
	node        Node
}

func NewContext() *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Reset()
	return ctx
}

func (ctx *Context) Reset() {
	ctx.Path = ""
	ctx.keys = ctx.keys[:0]
	ctx.vals = ctx.vals[:0]
	ctx.paramsCount = 0
	ctx.routerName = ""
	ctx.node = nil
}

func (ctx *Context) Params() Params { return ctx }

func (ctx *Context) SetNode(n Node) { ctx.node = n }

func (ctx *Context) SetRouterName(n string) { ctx.routerName = n }

func (ctx *Context) Node() Node { return ctx.node }

func (ctx *Context) RouterName() string { return ctx.routerName }

func (ctx *Context) Destroy() {
	const destroyMaxSize = 30
	if ctx != nil && len(ctx.keys) <= destroyMaxSize {
		contextPool.Put(ctx)
	}
}

func (ctx *Context) Exists(key string) bool {
	_, found := ctx.Get(key)
	return found
}

func (ctx *Context) String(key string) (string, error) {
	if v, found := ctx.Get(key); found {
		return v, nil
	}
	return "", ErrParamNotExists()
}

func (ctx *Context) MustString(key, def string) string {
	if v, found := ctx.Get(key); found {
		return v
	}
	return def
}

func (ctx *Context) Int(key string) (int64, error) {
	if str, found := ctx.Get(key); found {
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, ErrParamNotExists()
}

func (ctx *Context) MustInt(key string, def int64) int64 {
	if str, found := ctx.Get(key); found {
		if val, err := strconv.ParseInt(str, 10, 64); err == nil {
			return val
		}
	}
	return def
}

func (ctx *Context) Uint(key string) (uint64, error) {
	if str, found := ctx.Get(key); found {
		return strconv.ParseUint(str, 10, 64)
	}
	return 0, ErrParamNotExists()
}

func (ctx *Context) MustUint(key string, def uint64) uint64 {
	if str, found := ctx.Get(key); found {
		if val, err := strconv.ParseUint(str, 10, 64); err == nil {
			return val
		}
	}
	return def
}

func (ctx *Context) Bool(key string) (bool, error) {
	if str, found := ctx.Get(key); found {
		return strconv.ParseBool(str)
	}
	return false, ErrParamNotExists()
}

func (ctx *Context) MustBool(key string, def bool) bool {
	if str, found := ctx.Get(key); found {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return def
}

func (ctx *Context) Float(key string) (float64, error) {
	if str, found := ctx.Get(key); found {
		return strconv.ParseFloat(str, 64)
	}
	return 0, ErrParamNotExists()
}

func (ctx *Context) MustFloat(key string, def float64) float64 {
	if str, found := ctx.Get(key); found {
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return val
		}
	}
	return def
}

func (ctx *Context) Get(key string) (string, bool) {
	if ctx == nil {
		return "", false
	}

	for i, k := range ctx.keys {
		if k == key {
			return ctx.vals[i], true
		}
	}
	return "", false
}

func (ctx *Context) Count() (cnt int) {
	if ctx == nil {
		return 0
	}
	return ctx.paramsCount
}

func (ctx *Context) Set(k, v string) {
	deletedIndex := -1

	for i, key := range ctx.keys {
		if key == k {
			ctx.vals[i] = v
			return
		}
		if key == "" && deletedIndex == -1 {
			deletedIndex = i
		}
	}

	// 没有需要修改的项
	ctx.paramsCount++
	if deletedIndex != -1 { // 优先考虑被标记为删除的项作为添加
		ctx.keys[deletedIndex] = k
		ctx.vals[deletedIndex] = v
	} else {
		ctx.keys = append(ctx.keys, k)
		ctx.vals = append(ctx.vals, v)
	}
}

func (ctx *Context) Delete(k string) {
	if ctx == nil {
		return
	}

	for i, key := range ctx.keys {
		if key == k {
			ctx.keys[i] = ""
			ctx.paramsCount--
			return
		}
	}
}

func (ctx *Context) Range(f func(key, val string)) {
	for i, k := range ctx.keys {
		if k != "" {
			f(k, ctx.vals[i])
		}
	}
}
