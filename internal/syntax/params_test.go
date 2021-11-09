// SPDX-License-Identifier: MIT

package syntax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/mux/v5/params"
)

var _ params.Params = &Params{}

func getParams(params *Params, a *assert.Assertion) *Params {
	r := httptest.NewRequest(http.MethodGet, "/to/path", nil)
	r = WithValue(r, params)
	a.NotNil(r)
	return GetParams(r)
}

func TestWithValue(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/to/path", nil)
	a.Equal(WithValue(r, &Params{}), r)

	r = httptest.NewRequest(http.MethodGet, "/to/path", nil)
	r = WithValue(r, &Params{Params: []Param{{K: "k1", V: "v1"}}})
	r = WithValue(r, &Params{Params: []Param{{K: "k2", V: "v2"}}})
	a.Equal(GetParams(r), &Params{Params: []Param{{K: "k2", V: "v2"}, {K: "k1", V: "v1"}}})
}

func TestGetParams(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/to/path", nil)
	ps := GetParams(r)
	a.Nil(ps)

	kvs := []Param{{K: "key1", V: "1"}}
	r = httptest.NewRequest(http.MethodGet, "/to/path", nil)
	ctx := context.WithValue(r.Context(), contextKeyParams, &Params{Params: kvs})
	r = r.WithContext(ctx)
	ps = GetParams(r)
	a.Equal(ps.Params, kvs)
}

func TestNewParams(t *testing.T) {
	a := assert.New(t)

	var p *Params
	p.Destroy()

	p = NewParams("/abc")
	a.Equal(p.Path, "/abc")
	p.Destroy()

	p = NewParams("/def")
	a.Equal(p.Path, "/def")
}

func TestParams_String(t *testing.T) {
	a := assert.New(t)

	ps := getParams(&Params{Params: []Param{{K: "key1", V: "1"}}}, a)

	val, err := ps.String("key1")
	a.NotError(err).Equal(val, "1")
	a.True(ps.Exists("key1"))
	a.Equal(ps.MustString("key1", "-9"), "1")

	// 不存在
	val, err = ps.String("k5")
	a.ErrorType(err, params.ErrParamNotExists).Equal(val, "")
	a.False(ps.Exists("k5"))
	a.Equal(ps.MustString("k5", "-10"), "-10")
}

func TestParams_Int(t *testing.T) {
	a := assert.New(t)

	ps := getParams(&Params{Params: []Param{
		{K: "key1", V: "1"},
		{K: "key2", V: "a2"},
	}}, a)

	val, err := ps.Int("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(ps.MustInt("key1", -9), 1)

	// 无法转换
	val, err = ps.Int("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(ps.MustInt("key2", -9), -9)

	// 不存在
	val, err = ps.Int("k5")
	a.ErrorType(err, params.ErrParamNotExists).Equal(val, 0)
	a.Equal(ps.MustInt("k5", -10), -10)
}

func TestParams_Uint(t *testing.T) {
	a := assert.New(t)

	ps := getParams(&Params{Params: []Param{
		{K: "key1", V: "1"},
		{K: "key2", V: "a2"},
		{K: "key3", V: "-1"},
	}}, a)

	val, err := ps.Uint("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(ps.MustUint("key1", 9), 1)

	// 无法转换
	val, err = ps.Uint("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(ps.MustUint("key2", 9), 9)

	// 负数
	val, err = ps.Uint("key3")
	a.Error(err).Equal(val, 0)
	a.Equal(ps.MustUint("key3", 9), 9)

	// 不存在
	val, err = ps.Uint("k5")
	a.ErrorType(err, params.ErrParamNotExists).Equal(val, 0)
	a.Equal(ps.MustUint("k5", 10), 10)
}

func TestParams_Bool(t *testing.T) {
	a := assert.New(t)

	ps := getParams(&Params{Params: []Param{
		{K: "key1", V: "true"},
		{K: "key2", V: "0"},
		{K: "key3", V: "a3"},
	}}, a)

	val, err := ps.Bool("key1")
	a.NotError(err).True(val)
	a.True(ps.MustBool("key1", false))

	val, err = ps.Bool("key2")
	a.NotError(err).False(val)
	a.False(ps.MustBool("key2", true))

	// 无法转换
	val, err = ps.Bool("key3")
	a.Error(err).False(val)
	a.True(ps.MustBool("key3", true))

	// 不存在
	val, err = ps.Bool("k5")
	a.ErrorType(err, params.ErrParamNotExists).False(val)
	a.True(ps.MustBool("k5", true))
}

func TestParams_Float(t *testing.T) {
	a := assert.New(t)

	ps := getParams(&Params{Params: []Param{
		{K: "key1", V: "1"},
		{K: "key2", V: "a2"},
		{K: "key3", V: "1.1"},
	}}, a)

	val, err := ps.Float("key1")
	a.NotError(err).Equal(val, 1.0)
	a.Equal(ps.MustFloat("key1", -9.0), 1.0)

	val, err = ps.Float("key3")
	a.NotError(err).Equal(val, 1.1)
	a.Equal(ps.MustFloat("key3", -9.0), 1.1)

	// 无法转换
	val, err = ps.Float("key2")
	a.Error(err).Equal(val, 0.0)
	a.Equal(ps.MustFloat("key2", -9.0), -9.0)

	// 不存在
	val, err = ps.Float("k5")
	a.ErrorType(err, params.ErrParamNotExists).Equal(val, 0.0)
	a.Equal(ps.MustFloat("k5", -10.0), -10.0)

	var ps2 *Params
	val, err = ps2.Float("key1")
	a.Equal(err, params.ErrParamNotExists).Equal(val, 0.0)
}

func TestParams_Set(t *testing.T) {
	a := assert.New(t)

	ps := &Params{Params: []Param{{K: "k1", V: "v1"}}}
	a.Equal(ps.Count(), 1)

	ps.Set("k1", "v2")
	a.Equal(ps.Count(), 1)
	a.Equal(ps, &Params{Params: []Param{{K: "k1", V: "v2"}}})

	ps.Set("k2", "v2")
	a.Equal(ps, &Params{Params: []Param{{K: "k1", V: "v2"}, {K: "k2", V: "v2"}}})
	a.Equal(ps.Count(), 2)
}

func TestParams_Get(t *testing.T) {
	a := assert.New(t)

	var ps *Params
	a.Zero(ps.Count())
	v, found := ps.Get("not-exists")
	a.False(found).Zero(v)

	ps = &Params{Params: []Param{{K: "k1", V: "v1"}}}
	v, found = ps.Get("k1")
	a.True(found).Equal(v, "v1")

	v, found = ps.Get("not-exists")
	a.False(found).Zero(v)

}

func TestParams_Clone(t *testing.T) {
	a := assert.New(t)

	var ps *Params
	ps2 := ps.Clone()
	a.Nil(ps2)

	ps = NewParams("/path")
	ps.Set("k1", "v1")
	ps.Set("k2", "v2")
	ps2 = ps.Clone()
	a.Equal(ps2.MustString("k1", "invalid"), "v1").
		Equal(ps2.MustString("k2", "invalid"), "v2")
}

func TestParams_Delete(t *testing.T) {
	a := assert.New(t)

	var ps *Params
	ps.Delete("k1")

	ps = NewParams("/path")
	ps.Set("k1", "v1")
	ps.Set("k2", "v2")
	ps2 := ps.Clone()
	a.Equal(2, ps.Count()).
		Equal(2, ps2.Count())

	ps.Delete("k1")
	a.Equal(1, ps.Count()).
		Equal(2, ps2.Count())
	ps.Delete("k1") // 多次删除同一个值
	a.Equal(1, ps.Count()).
		Equal(2, ps2.Count())

	ps.Delete("k2")
	a.Equal(0, ps.Count()).
		Equal(2, ps2.Count()).
		Equal(2, len(ps.Params))

	ps.Set("k3", "v3")
	a.Equal(1, ps.Count()).
		Equal(2, ps2.Count()).
		Equal(2, len(ps.Params))
}

func TestParams_Map(t *testing.T) {
	a := assert.New(t)

	var ps *Params
	a.Nil(ps.Map())

	ps = NewParams("/path")
	a.Nil(ps.Map())

	ps.Set("k1", "v1")
	a.Equal(ps.Map(), map[string]string{"k1": "v1"})

	ps.Delete("k1")
	a.Empty(ps.Map())
}
