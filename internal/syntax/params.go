// SPDX-License-Identifier: MIT

package syntax

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/issue9/mux/v6/params"
)

// 每次申请 Params.Params 分配的大小
const defaultParamsCap = 5

var paramsPool = &sync.Pool{
	New: func() any { return &Params{Params: make([]Param, 0, defaultParamsCap)} },
}

const contextKeyParams contextKey = 0

type contextKey int

// Params 路由参数
//
// 实现了 params.Params 接口
type Params struct {
	Path   string  // 这是在 Segment.Match 中用到的路径信息。
	Params []Param // 实际需要传递的参数
}

type Param struct {
	K, V string // 如果 K 为空，则表示该参数已经被删除。
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
func WithValue(r *http.Request, ps params.Params) *http.Request {
	if ps == nil || ps.Count() == 0 {
		return r
	}

	pp := ps.(*Params)
	if ps2 := GetParams(r); ps2 != nil && len(ps2.Params) > 0 {
		for _, p := range ps2.Params {
			pp.Set(p.K, p.V)
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

func (p *Params) Map() map[string]string {
	if p == nil || len(p.Params) == 0 {
		return nil
	}

	m := make(map[string]string, len(p.Params))
	for _, pp := range p.Params {
		if pp.K != "" {
			m[pp.K] = pp.V
		}
	}
	return m
}

func (p *Params) Set(k, v string) {
	deletedIndex := -1

	for i, param := range p.Params {
		if param.K == k {
			p.Params[i].V = v
			return
		}
		if param.K == "" && deletedIndex == -1 {
			deletedIndex = i
		}
	}

	if deletedIndex != -1 {
		p.Params[deletedIndex].K = k
		p.Params[deletedIndex].V = v
	} else {
		p.Params = append(p.Params, Param{K: k, V: v})
	}

}

func (p *Params) Delete(k string) {
	if p == nil {
		return
	}

	for i, pp := range p.Params {
		if pp.K == k {
			p.Params[i].K = ""
			return
		}
	}
}
