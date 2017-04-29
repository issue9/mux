// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"context"
	"testing"

	"net/http"

	"github.com/issue9/assert"
)

func setParams(params map[string]string, a *assert.Assertion) *http.Request {
	r, err := http.NewRequest(http.MethodGet, "/to/path", nil)
	a.NotError(err).NotNil(r)

	ctx := context.WithValue(r.Context(), contextKeyParams, Params(params))
	return r.WithContext(ctx)
}

func TestGetParams(t *testing.T) {
	a := assert.New(t)

	r := setParams(nil, a)
	ps := GetParams(r)
	a.Nil(ps)

	r = setParams(map[string]string{"key1": "1"}, a)
	ps = GetParams(r)
	a.Equal(ps, map[string]string{"key1": "1"})
}

func TestParams_Int(t *testing.T) {
	a := assert.New(t)

	ps := Params{
		"key1": "1",
		"key2": "a2",
	}

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
