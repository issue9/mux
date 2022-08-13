// SPDX-License-Identifier: MIT

package types

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func BenchmarkContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := NewContext()
		ctx.Destroy()
	}
}

func BenchmarkContext_Get(b *testing.B) {
	a := assert.New(b, false)

	b.Run("1", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")
		for i := 0; i < b.N; i++ {
			_, found := ctx.Get("K1")
			a.True(found)
		}
	})

	b.Run("3", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")
		ctx.Set("K2", "v2")
		ctx.Set("K3", "v3")
		for i := 0; i < b.N; i++ {
			_, found := ctx.Get("K3")
			a.True(found)
		}
	})

	b.Run("5", func(b *testing.B) {
		ctx := NewContext()
		ctx.Set("K1", "v1")
		ctx.Set("K2", "v2")
		ctx.Set("K3", "v3")
		ctx.Set("K4", "v4")
		ctx.Set("K5", "v5")
		for i := 0; i < b.N; i++ {
			_, found := ctx.Get("K5")
			a.True(found)
		}
	})

	b.Run("10", func(b *testing.B) {
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
		for i := 0; i < b.N; i++ {
			_, found := ctx.Get("K10")
			a.True(found)
		}
	})
}
