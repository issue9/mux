// SPDX-License-Identifier: MIT

package syntax

import (
	"errors"
	"strconv"
	"sync"
)

var ErrParamNotExists = errors.New("不存在该参数")

var paramsPool = &sync.Pool{
	New: func() any { return &Params{Params: make([]Param, 0, 5)} },
}

// Params 路由参数
//
// 实现了 types.Params 接口
type Params struct {
	Path   string  // 这是在 Segment.Match 中用到的路径信息。
	Params []Param // 实际需要传递的参数
	count  int
}

type Param struct {
	K, V string // 如果 K 为空，则表示该参数已经被删除。
}

func NewParams(path string) *Params {
	ps := paramsPool.Get().(*Params)
	ps.Path = path
	ps.Params = ps.Params[:0]
	ps.count = 0
	return ps
}

func (p *Params) Destroy() {
	if p != nil {
		paramsPool.Put(p)
	}
}

func (p *Params) Exists(key string) bool {
	_, found := p.Get(key)
	return found
}

func (p *Params) String(key string) (string, error) {
	if v, found := p.Get(key); found {
		return v, nil
	}
	return "", ErrParamNotExists
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
	return 0, ErrParamNotExists
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
	return 0, ErrParamNotExists
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
	return false, ErrParamNotExists
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
	return 0, ErrParamNotExists
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

func (p *Params) Count() (cnt int) {
	if p == nil {
		return 0
	}
	return p.count
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

	p.count++
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
			p.count--
			return
		}
	}
}

func (p *Params) Range(f func(key, val string)) {
	for _, param := range p.Params {
		if param.K != "" {
			f(param.K, param.V)
		}
	}
}
