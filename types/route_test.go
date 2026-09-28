// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"testing"

	"github.com/issue9/assert/v5"
)

func TestNewContext(t *testing.T) {
	a := assert.New(t, false)

	var p *Route
	p.Destroy()

	p = NewRoute()
	p.Path = "/abc"
	a.Equal(p.Path, "/abc")
	p.Destroy()

	p = NewRoute()
	p.Path = "/def"
	a.Equal(p.Path, "/def")
}

func TestContext_String(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("key1", "1")

	val, err := route.String("key1")
	a.NotError(err).Equal(val, "1")
	a.True(route.Exists("key1"))
	a.Equal(route.MustString("key1", "-9"), "1")

	// 不存在
	val, err = route.String("k5")
	a.ErrorIs(err, ErrParamNotExists()).Equal(val, "")
	a.False(route.Exists("k5"))
	a.Equal(route.MustString("k5", "-10"), "-10")
}

func TestContext_Int(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("key1", "1")
	route.Set("key2", "a2")

	val, err := route.Int("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(route.MustInt("key1", -9), 1)

	// 无法转换
	val, err = route.Int("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(route.MustInt("key2", -9), -9)

	// 不存在
	val, err = route.Int("k5")
	a.ErrorIs(err, ErrParamNotExists()).Equal(val, 0)
	a.Equal(route.MustInt("k5", -10), -10)
}

func TestContext_Uint(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("key1", "1")
	route.Set("key2", "a2")
	route.Set("key3", "-1")

	val, err := route.Uint("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(route.MustUint("key1", 9), 1)

	// 无法转换
	val, err = route.Uint("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(route.MustUint("key2", 9), 9)

	// 负数
	val, err = route.Uint("key3")
	a.Error(err).Equal(val, 0)
	a.Equal(route.MustUint("key3", 9), 9)

	// 不存在
	val, err = route.Uint("k5")
	a.ErrorIs(err, ErrParamNotExists()).Equal(val, 0)
	a.Equal(route.MustUint("k5", 10), 10)
}

func TestContext_Bool(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("key1", "true")
	route.Set("key2", "0")
	route.Set("key3", "a3")

	val, err := route.Bool("key1")
	a.NotError(err).True(val)
	a.True(route.MustBool("key1", false))

	val, err = route.Bool("key2")
	a.NotError(err).False(val)
	a.False(route.MustBool("key2", true))

	// 无法转换
	val, err = route.Bool("key3")
	a.Error(err).False(val)
	a.True(route.MustBool("key3", true))

	// 不存在
	val, err = route.Bool("k5")
	a.ErrorIs(err, ErrParamNotExists()).False(val)
	a.True(route.MustBool("k5", true))
}

func TestContext_Float(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("key1", "1")
	route.Set("key2", "a2")
	route.Set("key3", "1.1")

	val, err := route.Float("key1")
	a.NotError(err).Equal(val, 1.0)
	a.Equal(route.MustFloat("key1", -9.0), 1.0)

	val, err = route.Float("key3")
	a.NotError(err).Equal(val, 1.1)
	a.Equal(route.MustFloat("key3", -9.0), 1.1)

	// 无法转换
	val, err = route.Float("key2")
	a.Error(err).Equal(val, 0.0)
	a.Equal(route.MustFloat("key2", -9.0), -9.0)

	// 不存在
	val, err = route.Float("k5")
	a.ErrorIs(err, ErrParamNotExists()).Equal(val, 0.0)
	a.Equal(route.MustFloat("k5", -10.0), -10.0)
}

func TestContext_Set(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("k1", "v1")
	a.Equal(route.Count(), 1)

	route.Set("k1", "v2")
	a.Equal(route.Count(), 1).
		Equal(route.params, map[string]string{"k1": "v2"})

	route.Set("k2", "v2")
	a.Equal(route.params, map[string]string{"k1": "v2", "k2": "v2"}).
		Equal(route.Count(), 2)
}

func TestContext_Get(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Set("k1", "v1")
	v, found := route.Get("k1")
	a.True(found).Equal(v, "v1")

	v, found = route.Get("not-exists")
	a.False(found).Zero(v)

}

func TestContext_Delete(t *testing.T) {
	a := assert.New(t, false)

	route := NewRoute()
	route.Path = "/path"
	route.Set("k1", "v1")
	route.Set("k2", "v2")

	route.Delete("k1")
	a.Equal(1, route.Count())
	route.Delete("k1") // 多次删除同一个值
	a.Equal(1, route.Count())

	route.Delete("k2")
	a.Equal(0, route.Count())

	route.Set("k3", "v3")
	a.Equal(1, route.Count())
}

func TestContext_IterSeq(t *testing.T) {
	a := assert.New(t, false)
	var size int

	ps := NewRoute()
	ps.Path = "/path"
	ps.Set("k1", "v1")
	ps.Set("k2", "v2")
	for range ps.Params() {
		size++
	}
	a.Equal(2, size)
}
