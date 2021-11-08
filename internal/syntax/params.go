// SPDX-License-Identifier: MIT

package syntax

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/issue9/mux/v5/params"
)

// 每次申请 Params.Params 分配的大小
const defaultParamsCap = 5

var paramsPool = &sync.Pool{
	New: func() interface{} { return &Params{Params: make([]Param, 0, defaultParamsCap)} },
}

const contextKeyParams contextKey = 0

type contextKey int

// Params 路由参数
//
// 实现了 mux.Params 接口
type Params struct {
	Path   string  // 这是在 Segment.Match 中用到的路径信息。
	Params []Param // 实际需要传递的参数
}

type Param struct {
	K, V string
}

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

func (p *Params) Exists(key string) bool {
	_, found := p.Get(key)
	return found
}

func (p *Params) String(key string) (string, error) {
	if v, found := p.Get(key); found {
		return v, nil
	}
	return "", params.ErrParamNotExists
}

func (p *Params) MustString(key, def string) string {
	if v, found := p.Get(key); found {
		return v
	}
	return def
}

func (p *Params) Int(key string) (int64, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, params.ErrParamNotExists
}

func (p *Params) MustInt(key string, def int64) int64 {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseInt(str, 10, 64); err == nil {
			return val
		}
	}
	return def
}

func (p *Params) Uint(key string) (uint64, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseUint(str, 10, 64)
	}
	return 0, params.ErrParamNotExists
}

func (p *Params) MustUint(key string, def uint64) uint64 {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseUint(str, 10, 64); err == nil {
			return val
		}
	}
	return def
}

func (p *Params) Bool(key string) (bool, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseBool(str)
	}
	return false, params.ErrParamNotExists
}

func (p *Params) MustBool(key string, def bool) bool {
	if str, found := p.Get(key); found {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return def
}

func (p *Params) Float(key string) (float64, error) {
	if str, found := p.Get(key); found {
		return strconv.ParseFloat(str, 64)
	}
	return 0, params.ErrParamNotExists
}

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

func (p *Params) Clone() params.Params {
	if p == nil {
		return nil
	}
	pp := &Params{
		Path:   p.Path,
		Params: make([]Param, len(p.Params)),
	}
	copy(pp.Params, p.Params)
	return pp
}

func (p *Params) Count() (cnt int) {
	if p == nil {
		return 0
	}

	for _, param := range p.Params {
		if param.K != "" {
			cnt++
		}
	}
	return cnt
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
	if p == nil {
		return
	}

	for i, pp := range p.Params {
		if pp.K == k {
			p.Params[i] = Param{}
			return
		}
	}
}
