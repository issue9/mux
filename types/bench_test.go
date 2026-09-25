// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"strconv"
	"testing"

	"github.com/issue9/assert/v5"
)

func BenchmarkContext(b *testing.B) {
	b.Run("no destroy", func(b *testing.B) {
		for range b.N {
			NewContext()
		}
	})

	b.Run("destroy", func(b *testing.B) {
		for range b.N {
			NewContext().Destroy()
		}
	})
}

func BenchmarkContext_Get(b *testing.B) {
	a := assert.New(b, false)

	b.Run("1 param", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")

		var found bool
		for b.Loop() {
			_, found = ctx.Get("K1")
		}
		a.True(found) // 结果是否正确由测试代码保证，性能测试只是检测最后一次的结果是否正确。

		ctx.Destroy()
	})

	b.Run("5 param", func(b *testing.B) {
		ctx := NewContext()
		for i := range 5 {
			ctx.Set("K"+strconv.Itoa(i+1), "v"+strconv.Itoa(i+1))
		}

		var found bool
		for b.Loop() {
			_, found = ctx.Get("K5")
		}
		a.True(found)

		ctx.Destroy()
	})

	b.Run("10 param", func(b *testing.B) {
		ctx := NewContext()
		for i := range 10 {
			ctx.Set("K"+strconv.Itoa(i+1), "v"+strconv.Itoa(i+1))
		}

		var found bool
		for b.Loop() {
			_, found = ctx.Get("K10")
		}
		a.True(found)

		ctx.Destroy()
	})
}
