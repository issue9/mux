// SPDX-License-Identifier: MIT

package syntax

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v6/types"
)

var _ types.Params = &Params{}

func TestNewParams(t *testing.T) {
	a := assert.New(t, false)

	var p *Params
	p.Destroy()

	p = NewParams("/abc")
	a.Equal(p.Path, "/abc")
	p.Destroy()

	p = NewParams("/def")
	a.Equal(p.Path, "/def")
}

func TestParams_String(t *testing.T) {
	a := assert.New(t, false)

	ps := &Params{Params: []Param{{K: "key1", V: "1"}}}

	val, err := ps.String("key1")
	a.NotError(err).Equal(val, "1")
	a.True(ps.Exists("key1"))
	a.Equal(ps.MustString("key1", "-9"), "1")

	// 不存在
	val, err = ps.String("k5")
	a.ErrorIs(err, ErrParamNotExists).Equal(val, "")
	a.False(ps.Exists("k5"))
	a.Equal(ps.MustString("k5", "-10"), "-10")
}

func TestParams_Int(t *testing.T) {
	a := assert.New(t, false)

	ps := &Params{Params: []Param{
		{K: "key1", V: "1"},
		{K: "key2", V: "a2"},
	}}

	val, err := ps.Int("key1")
	a.NotError(err).Equal(val, 1)
	a.Equal(ps.MustInt("key1", -9), 1)

	// 无法转换
	val, err = ps.Int("key2")
	a.Error(err).Equal(val, 0)
	a.Equal(ps.MustInt("key2", -9), -9)

	// 不存在
	val, err = ps.Int("k5")
	a.ErrorIs(err, ErrParamNotExists).Equal(val, 0)
	a.Equal(ps.MustInt("k5", -10), -10)
}

func TestParams_Uint(t *testing.T) {
	a := assert.New(t, false)

	ps := &Params{Params: []Param{
		{K: "key1", V: "1"},
		{K: "key2", V: "a2"},
		{K: "key3", V: "-1"},
	}}

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
	a.ErrorIs(err, ErrParamNotExists).Equal(val, 0)
	a.Equal(ps.MustUint("k5", 10), 10)
}

func TestParams_Bool(t *testing.T) {
	a := assert.New(t, false)

	ps := &Params{Params: []Param{
		{K: "key1", V: "true"},
		{K: "key2", V: "0"},
		{K: "key3", V: "a3"},
	}}

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
	a.ErrorIs(err, ErrParamNotExists).False(val)
	a.True(ps.MustBool("k5", true))
}

func TestParams_Float(t *testing.T) {
	a := assert.New(t, false)

	ps := &Params{Params: []Param{
		{K: "key1", V: "1"},
		{K: "key2", V: "a2"},
		{K: "key3", V: "1.1"},
	}}

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
	a.ErrorIs(err, ErrParamNotExists).Equal(val, 0.0)
	a.Equal(ps.MustFloat("k5", -10.0), -10.0)

	var ps2 *Params
	val, err = ps2.Float("key1")
	a.Equal(err, ErrParamNotExists).Equal(val, 0.0)
}

func TestParams_Set(t *testing.T) {
	a := assert.New(t, false)

	ps := NewParams("")
	ps.Set("k1", "v1")
	a.Equal(ps.Count(), 1)

	ps.Set("k1", "v2")
	a.Equal(ps.Count(), 1)
	a.Equal(ps, &Params{Params: []Param{{K: "k1", V: "v2"}}, count: 1})

	ps.Set("k2", "v2")
	a.Equal(ps, &Params{Params: []Param{{K: "k1", V: "v2"}, {K: "k2", V: "v2"}}, count: 2})
	a.Equal(ps.Count(), 2)
}

func TestParams_Get(t *testing.T) {
	a := assert.New(t, false)

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

func TestParams_Delete(t *testing.T) {
	a := assert.New(t, false)

	var ps *Params
	ps.Delete("k1")

	ps = NewParams("/path")
	ps.Set("k1", "v1")
	ps.Set("k2", "v2")

	ps.Delete("k1")
	a.Equal(1, ps.Count())
	ps.Delete("k1") // 多次删除同一个值
	a.Equal(1, ps.Count())

	ps.Delete("k2")
	a.Equal(0, ps.Count()).
		Equal(2, len(ps.Params))

	ps.Set("k3", "v3")
	a.Equal(1, ps.Count()).
		Equal(2, len(ps.Params))
}

func TestParams_Range(t *testing.T) {
	a := assert.New(t, false)
	var size int

	ps := NewParams("/path")
	ps.Set("k1", "v1")
	ps.Set("k2", "v2")
	ps.Range(func(k, v string) {
		size++
	})
	a.Equal(2, size)
}
