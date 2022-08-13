// SPDX-License-Identifier: MIT

package types

import (
	"testing"

	"github.com/issue9/assert/v3"
)

var _ Route = &Context{}

func TestNewContext(t *testing.T) {
	a := assert.New(t, false)

	var p *Context
	p.Destroy()

	p = NewContext()
	p.Path = "/abc"
	a.Equal(p.Path, "/abc")
	p.Destroy()

	p = NewContext()
	p.Path = "/def"
	a.Equal(p.Path, "/def")
}

func TestContext_String(t *testing.T) {
	a := assert.New(t, false)

	ctx := NewContext()
	ctx.Set("key1", "1")

	val, err := ctx.String("key1")
	a.NotError(err).Equal(val, "1")
	a.True(ctx.Exists("key1"))
	a.Equal(ctx.MustString("key1", "-9"), "1")

	// 不存在
	val, err = ctx.String("k5")
	a.ErrorIs(err, ErrParamNotExists).Equal(val, "")
	a.False(ctx.Exists("k5"))
	a.Equal(ctx.MustString("k5", "-10"), "-10")
}

func TestContext_Int(t *testing.T) {
	a := assert.New(t, false)

	ctx := NewContext()
	ctx.Set("key1", "1")
	ctx.Set("key2", "a2")

	val, err := ctx.Int("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(ctx.MustInt("key1", -9), 1)

	// 无法转换
	val, err = ctx.Int("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(ctx.MustInt("key2", -9), -9)

	// 不存在
	val, err = ctx.Int("k5")
	a.ErrorIs(err, ErrParamNotExists).Equal(val, 0)
	a.Equal(ctx.MustInt("k5", -10), -10)
}

func TestContext_Uint(t *testing.T) {
	a := assert.New(t, false)

	ctx := NewContext()
	ctx.Set("key1", "1")
	ctx.Set("key2", "a2")
	ctx.Set("key3", "-1")

	val, err := ctx.Uint("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(ctx.MustUint("key1", 9), 1)

	// 无法转换
	val, err = ctx.Uint("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(ctx.MustUint("key2", 9), 9)

	// 负数
	val, err = ctx.Uint("key3")
	a.Error(err).Equal(val, 0)
	a.Equal(ctx.MustUint("key3", 9), 9)

	// 不存在
	val, err = ctx.Uint("k5")
	a.ErrorIs(err, ErrParamNotExists).Equal(val, 0)
	a.Equal(ctx.MustUint("k5", 10), 10)
}

func TestContext_Bool(t *testing.T) {
	a := assert.New(t, false)

	ctx := NewContext()
	ctx.Set("key1", "true")
	ctx.Set("key2", "0")
	ctx.Set("key3", "a3")

	val, err := ctx.Bool("key1")
	a.NotError(err).True(val)
	a.True(ctx.MustBool("key1", false))

	val, err = ctx.Bool("key2")
	a.NotError(err).False(val)
	a.False(ctx.MustBool("key2", true))

	// 无法转换
	val, err = ctx.Bool("key3")
	a.Error(err).False(val)
	a.True(ctx.MustBool("key3", true))

	// 不存在
	val, err = ctx.Bool("k5")
	a.ErrorIs(err, ErrParamNotExists).False(val)
	a.True(ctx.MustBool("k5", true))
}

func TestContext_Float(t *testing.T) {
	a := assert.New(t, false)

	ctx := NewContext()
	ctx.Set("key1", "1")
	ctx.Set("key2", "a2")
	ctx.Set("key3", "1.1")

	val, err := ctx.Float("key1")
	a.NotError(err).Equal(val, 1.0)
	a.Equal(ctx.MustFloat("key1", -9.0), 1.0)

	val, err = ctx.Float("key3")
	a.NotError(err).Equal(val, 1.1)
	a.Equal(ctx.MustFloat("key3", -9.0), 1.1)

	// 无法转换
	val, err = ctx.Float("key2")
	a.Error(err).Equal(val, 0.0)
	a.Equal(ctx.MustFloat("key2", -9.0), -9.0)

	// 不存在
	val, err = ctx.Float("k5")
	a.ErrorIs(err, ErrParamNotExists).Equal(val, 0.0)
	a.Equal(ctx.MustFloat("k5", -10.0), -10.0)

	var ps2 *Context
	val, err = ps2.Float("key1")
	a.Equal(err, ErrParamNotExists).Equal(val, 0.0)
}

func TestContext_Set(t *testing.T) {
	a := assert.New(t, false)

	ctx := NewContext()
	ctx.Set("k1", "v1")
	a.Equal(ctx.Count(), 1)

	ctx.Set("k1", "v2")
	a.Equal(ctx.Count(), 1)
	a.Equal(ctx.keys, []string{"k1"})
	a.Equal(ctx.vals, []string{"v2"})

	ctx.Set("k2", "v2")
	a.Equal(ctx.keys, []string{"k1", "k2"})
	a.Equal(ctx.vals, []string{"v2", "v2"})
	a.Equal(ctx.Count(), 2)
}

func TestContext_Get(t *testing.T) {
	a := assert.New(t, false)

	var ctx *Context
	a.Zero(ctx.Count())
	v, found := ctx.Get("not-exists")
	a.False(found).Zero(v)

	ctx = NewContext()
	ctx.Set("k1", "v1")
	v, found = ctx.Get("k1")
	a.True(found).Equal(v, "v1")

	v, found = ctx.Get("not-exists")
	a.False(found).Zero(v)

}

func TestContext_Delete(t *testing.T) {
	a := assert.New(t, false)

	var ctx *Context
	ctx.Delete("k1")

	ctx = NewContext()
	ctx.Path = "/path"
	ctx.Set("k1", "v1")
	ctx.Set("k2", "v2")

	ctx.Delete("k1")
	a.Equal(1, ctx.Count())
	ctx.Delete("k1") // 多次删除同一个值
	a.Equal(1, ctx.Count())

	ctx.Delete("k2")
	a.Equal(0, ctx.Count()).
		Equal(2, len(ctx.keys))

	ctx.Set("k3", "v3")
	a.Equal(1, ctx.Count()).
		Equal(2, len(ctx.keys))
}

func TestContext_Range(t *testing.T) {
	a := assert.New(t, false)
	var size int

	ps := NewContext()
	ps.Path = "/path"
	ps.Set("k1", "v1")
	ps.Set("k2", "v2")
	ps.Range(func(k, v string) {
		size++
	})
	a.Equal(2, size)
}
