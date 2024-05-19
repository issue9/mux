// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"testing"

	"github.com/issue9/assert/v4"
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
		for range b.N {
			_, found := ctx.Get("K1")
			a.True(found)
		}
		ctx.Destroy()
	})

	b.Run("3 param", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")
		ctx.Set("K2", "v2")
		ctx.Set("K3", "v3")
		for range b.N {
			_, found := ctx.Get("K3")
			a.True(found)
		}
		ctx.Destroy()
	})

	b.Run("5 param", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")
		ctx.Set("K2", "v2")
		ctx.Set("K3", "v3")
		ctx.Set("K4", "v4")
		ctx.Set("K5", "v5")
		for range b.N {
			_, found := ctx.Get("K5")
			a.True(found)
		}
		ctx.Destroy()
	})

	b.Run("10 param", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")
		ctx.Set("K2", "v2")
		ctx.Set("K3", "v3")
		ctx.Set("K4", "v4")
		ctx.Set("K5", "v5")
		ctx.Set("K6", "v6")
		ctx.Set("K7", "v7")
		ctx.Set("K8", "v8")
		ctx.Set("K9", "v9")
		ctx.Set("K10", "v10")
		for range b.N {
			_, found := ctx.Get("K10")
			a.True(found)
		}
		ctx.Destroy()
	})
}
