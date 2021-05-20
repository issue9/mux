// SPDX-License-Identifier: MIT

package params

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func getParams(params map[string]string, a *assert.Assertion) Params {
	r := httptest.NewRequest(http.MethodGet, "/to/path", nil)
	r = WithValue(r, params)
	return Get(r)
}

func TestWithValue(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/to/path", nil)
	a.Equal(WithValue(r, Params{}), r)

	r = httptest.NewRequest(http.MethodGet, "/to/path", nil)
	r = WithValue(r, Params{"k1": "v1"})
	r = WithValue(r, map[string]string{"k2": "v2"})
	a.Equal(Get(r), map[string]string{"k1": "v1", "k2": "v2"})
}

func TestGet(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/to/path", nil)
	ps := Get(r)
	a.Nil(ps)

	maps := map[string]string{"key1": "1"}
	r = httptest.NewRequest(http.MethodGet, "/to/path", nil)
	ctx := context.WithValue(r.Context(), ContextKeyParams, Params(maps))
	r = r.WithContext(ctx)
	ps = Get(r)
	a.Equal(ps, maps)
}

func TestParams_String(t *testing.T) {
	a := assert.New(t)

	ps := getParams(map[string]string{
		"key1": "1",
	}, a)

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

	ps := getParams(map[string]string{
		"key1": "1",
		"key2": "a2",
	}, a)

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

	ps := getParams(map[string]string{
		"key1": "1",
		"key2": "a2",
		"key3": "-1",
	}, a)

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

	ps := getParams(map[string]string{
		"key1": "true",
		"key2": "0",
		"key3": "a3",
	}, a)

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

	ps := getParams(map[string]string{
		"key1": "1",
		"key2": "a2",
		"key3": "1.1",
	}, a)

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
