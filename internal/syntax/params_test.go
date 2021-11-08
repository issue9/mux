// SPDX-License-Identifier: MIT

package syntax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

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

func TestParams_String(t *testing.T) {
	a := assert.New(t)

	ps := getParams(&Params{Params: []Param{{K: "key1", V: "1"}}}, a)

	val, err := ps.String("key1")
	a.NotError(err).Equal(val, "1")
	a.True(ps.Exists("key1"))
	a.Equal(ps.MustString("key1", "-9"), "1")

	// 不存在
	val, err = ps.String("k5")
	a.ErrorType(err, ErrParamNotExists).Equal(val, "")
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
	a.ErrorType(err, ErrParamNotExists).Equal(val, 0)
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
	a.ErrorType(err, ErrParamNotExists).Equal(val, 0)
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
	a.ErrorType(err, ErrParamNotExists).False(val)
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
	a.ErrorType(err, ErrParamNotExists).Equal(val, 0.0)
	a.Equal(ps.MustFloat("k5", -10.0), -10.0)
}

func TestParams_Set(t *testing.T) {
	a := assert.New(t)

	ps := &Params{Params: []Param{{K: "k1", V: "v1"}}}

	ps.Set("k1", "v2")
	a.Equal(ps, &Params{Params: []Param{{K: "k1", V: "v2"}}})

	ps.Set("k2", "v2")
	a.Equal(ps, &Params{Params: []Param{{K: "k1", V: "v2"}, {K: "k2", V: "v2"}}})
}
