// SPDX-License-Identifier: MIT

package syntax

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"github.com/issue9/sliceutil"
)

// 每次申请 params.Params 分配的大小
const defaultParamsCap = 5

var paramsPool = &sync.Pool{
	New: func() interface{} { return &Params{Params: make([]Param, 0, defaultParamsCap)} },
}

// ErrParamNotExists 表示地址参数中并不存在该名称的值
var ErrParamNotExists = errors.New("不存在该参数")

const contextKeyParams contextKey = 0

type contextKey int

type Params struct {
	Path   string
	Params []Param
}

type Param struct{ K, V string }

func NewParams(path string) *Params {
	ps := paramsPool.Get().(*Params)
	ps.Path = path
	ps.Params = ps.Params[:0]
	return ps
}

func (p *Params) Destroy() {
	if p != nil {
		paramsPool.Put(p)
	}
}

// GetParams 获取当前请求实例上的参数列表
func GetParams(r *http.Request) *Params {
	if ps := r.Context().Value(contextKeyParams); ps != nil {
		return ps.(*Params)
	}
	return nil
}

// WithValue 将参数 ps 附加在 r 上
//
// 与 context.WithValue 功能相同，但是考虑了在同一个 r 上调用多次 WithValue 的情况。
func WithValue(r *http.Request, ps *Params) *http.Request {
	if ps == nil || len(ps.Params) == 0 {
		return r
	}

	if ps2 := GetParams(r); ps2 != nil && len(ps2.Params) > 0 {
		for _, p := range ps2.Params {
			ps.Set(p.K, p.V)
		}
	}

	return r.WithContext(context.WithValue(r.Context(), contextKeyParams, ps))
}

// Exists 查找指定名称的参数是否存在
//
// NOTE: 如果是可选参数，可能会不存在。
func (p *Params) Exists(key string) bool {
	_, found := p.Get(key)
	return found
}

// String 获取地址参数中的名为 key 的变量并将其转换成 string
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p *Params) String(key string) (string, error) {
	if v, found := p.Get(key); found {
		return v, nil
	}
	return "", ErrParamNotExists
}

// MustString 获取地址参数中的名为 key 的变量并将其转换成 string
//
// 若不存在或是无法转换则返回 def。
func (p *Params) MustString(key, def string) string {
	if v, found := p.Get(key); found {
		return v
	}
	return def
}

// Int 获取地址参数中的名为 key 的变量并将其转换成 int64
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p *Params) Int(key string) (int64, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, ErrParamNotExists
}

// MustInt 获取地址参数中的名为 key 的变量并将其转换成 int64
//
// 若不存在或是无法转换则返回 def。
func (p *Params) MustInt(key string, def int64) int64 {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseInt(str, 10, 64); err == nil {
			return val
		}
	}
	return def
}

// Uint 获取地址参数中的名为 key 的变量并将其转换成 uint64
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p *Params) Uint(key string) (uint64, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseUint(str, 10, 64)
	}
	return 0, ErrParamNotExists
}

// MustUint 获取地址参数中的名为 key 的变量并将其转换成 uint64
//
// 若不存在或是无法转换则返回 def。
func (p *Params) MustUint(key string, def uint64) uint64 {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseUint(str, 10, 64); err == nil {
			return val
		}
	}
	return def
}

// Bool 获取地址参数中的名为 key 的变量并将其转换成 bool
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p *Params) Bool(key string) (bool, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseBool(str)
	}
	return false, ErrParamNotExists
}

// MustBool 获取地址参数中的名为 key 的变量并将其转换成 bool
//
// 若不存在或是无法转换则返回 def。
func (p Params) MustBool(key string, def bool) bool {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return def
}

// Float 获取地址参数中的名为 key 的变量并将其转换成 Float64
//
// 当参数不存在时，返回 ErrParamNotExists 错误。
func (p Params) Float(key string) (float64, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseFloat(str, 64)
	}
	return 0, ErrParamNotExists
}

// MustFloat 获取地址参数中的名为 key 的变量并将其转换成 float64
//
// 若不存在或是无法转换则返回 def。
func (p *Params) MustFloat(key string, def float64) float64 {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return val
		}
	}
	return def
}

func (p *Params) Get(key string) (string, bool) {
	if p == nil {
		return "", false
	}

	for _, kv := range p.Params {
		if kv.K == key {
			return kv.V, true
		}
	}
	return "", false
}

func (p *Params) Set(k, v string) {
	for i, param := range p.Params {
		if param.K == k {
			p.Params[i] = Param{K: k, V: v}
			return
		}
	}

	p.Params = append(p.Params, Param{
		K: k,
		V: v,
	})
}

func (p *Params) Delete(k string) {
	index := sliceutil.QuickDelete(p.Params, func(i int) bool {
		return p.Params[i].K == k
	})
	p.Params = p.Params[:index]
}
