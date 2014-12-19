// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestContext(t *testing.T) {
	a := assert.New(t)

	req1, err := http.NewRequest("GET", "/abc/", nil)
	a.NotError(err)
	a.NotNil(req1)

	ctx1 := GetContext(req1)

	ctx1.Set("key", "val")
	a.Equal(ctx1.MustGet("key", "default").(string), "val")

	ctx2 := GetContext(req1)
	a.Equal(ctx2.MustGet("key", "default").(string), "val")

	freeContext(req1)

	ctx3 := GetContext(req1)
	a.Equal(ctx3.MustGet("key", "default").(string), "default")
}

func BenchmarkContextWithPool(b *testing.B) {
	req, _ := http.NewRequest("GET", "/abc/", nil)
	for i := 0; i < b.N; i++ {
		ctx := GetContext(req)
		ctx.Set("abc", "abc")
		freeContext(req)
	}
}

func BenchmarkContextWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ctx := &context{items: make(map[interface{}]interface{})}
		ctx.Set("abc", "abc")
	}
}
