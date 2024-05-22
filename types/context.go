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
	Path       string // 实际请求的路径信息
	params     map[string]string
	routerName string
	node       Node
}

func NewContext() *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Reset()
	return ctx
}

func (ctx *Context) Reset() {
	ctx.Path = ""
	clear(ctx.params)
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
	if ctx != nil && len(ctx.params) <= destroyMaxSize {
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
	if ctx.params == nil {
		return "", false
	}
	v, f := ctx.params[key]
	return v, f
}

func (ctx *Context) Count() int { return len(ctx.params) }

func (ctx *Context) Set(k, v string) {
	if ctx.params == nil {
		ctx.params = map[string]string{k: v}
		return
	}
	ctx.params[k] = v
}

func (ctx *Context) Delete(k string) {
	if ctx.params != nil {
		delete(ctx.params, k)
	}
}

func (ctx *Context) Range(f func(key, val string)) {
	for k, v := range ctx.params {
		f(k, v)
	}
}
