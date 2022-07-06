// SPDX-License-Identifier: MIT

package types

import (
	"strconv"
	"sync"
)

var contextPool = &sync.Pool{
	New: func() any { return &Context{params: make([]param, 0, 5)} },
}

// Context 保存着路由匹配过程中的上下文关系
type Context struct {
	Path           string  // 实际请求的路径信息
	params         []param // 实际需要传递的参数
	parameterCount int
	routerName     string
	node           Node
}

type param struct {
	K, V string // 如果 K 为空，则表示该参数已经被删除。
}

func NewContext(path string) *Context {
	ps := contextPool.Get().(*Context)
	ps.Path = path
	ps.params = ps.params[:0]
	ps.parameterCount = 0
	ps.routerName = ""
	ps.node = nil
	return ps
}

func (ctx *Context) Params() Params { return ctx }

func (ctx *Context) SetNode(n Node) { ctx.node = n }

func (ctx *Context) SetRouterName(n string) { ctx.routerName = n }

func (ctx *Context) Node() Node { return ctx.node }

func (ctx *Context) RouterName() string { return ctx.routerName }

func (ctx *Context) Destroy() {
	if ctx != nil {
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
	return "", ErrParamNotExists
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
	return 0, ErrParamNotExists
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
	return 0, ErrParamNotExists
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
	return false, ErrParamNotExists
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
	return 0, ErrParamNotExists
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

	for _, kv := range ctx.params {
		if kv.K == key {
			return kv.V, true
		}
	}
	return "", false
}

func (ctx *Context) Count() (cnt int) {
	if ctx == nil {
		return 0
	}
	return ctx.parameterCount
}

func (ctx *Context) Set(k, v string) {
	deletedIndex := -1

	for i, param := range ctx.params {
		if param.K == k {
			ctx.params[i].V = v
			return
		}
		if param.K == "" && deletedIndex == -1 {
			deletedIndex = i
		}
	}

	// 没有需要修改的项
	ctx.parameterCount++
	if deletedIndex != -1 { // 优先考虑被标记为删除的项作为添加
		ctx.params[deletedIndex].K = k
		ctx.params[deletedIndex].V = v
	} else {
		ctx.params = append(ctx.params, param{K: k, V: v})
	}
}

func (ctx *Context) Delete(k string) {
	if ctx == nil {
		return
	}

	for i, pp := range ctx.params {
		if pp.K == k {
			ctx.params[i].K = ""
			ctx.parameterCount--
			return
		}
	}
}

func (ctx *Context) Range(f func(key, val string)) {
	for _, param := range ctx.params {
		if param.K != "" {
			f(param.K, param.V)
		}
	}
}
