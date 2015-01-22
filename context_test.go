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
	ok := ctx1.Add("key", "val")
	a.True(ok).Equal(ctx1.MustGet("key", "default").(string), "val")

	// 添加一个相同的值，失败
	a.False(ctx1.Add("key", "val1"))
	a.Equal(len(ctx1.items), 1)

	// 测试set
	ctx1.Set("key2", "val")
	ctx1.Set("key2", "val2")
	v, found := ctx1.Get("key2")
	a.True(found).Equal(v, "val2")

	// 同一个Request，应该是相同的值
	ctx2 := GetContext(req1)
	a.Equal(ctx2.MustGet("key", "default").(string), "val")

	FreeContext(req1)

	ctx3 := GetContext(req1)
	a.Equal(ctx3.MustGet("key", "default").(string), "default")
}
