// SPDX-License-Identifier: MIT

package mux

import (
	"context"
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/mux/v6/internal/syntax"
)

func TestWithValue(t *testing.T) {
	a := assert.New(t, false)

	r, err := http.NewRequest(http.MethodGet, "/to/path", nil)
	a.NotError(err).NotNil(r)
	a.Equal(WithValue(r, &syntax.Params{}), r)

	r, err = http.NewRequest(http.MethodGet, "/to/path", nil)
	a.NotError(err).NotNil(r)
	pp := syntax.NewParams("")
	pp.Set("k1", "v1")
	r = WithValue(r, pp)

	pp = syntax.NewParams("")
	pp.Set("k2", "v2")
	r = WithValue(r, pp)
	ps := GetParams(r)
	a.NotNil(ps).
		Equal(ps.MustString("k2", "def"), "v2").
		Equal(ps.MustString("k1", "def"), "v1")
}

func TestGetParams(t *testing.T) {
	a := assert.New(t, false)

	r, err := http.NewRequest(http.MethodGet, "/to/path", nil)
	a.NotError(err).NotNil(r)
	ps := GetParams(r)
	a.Nil(ps)

	kvs := []syntax.Param{{K: "key1", V: "1"}}
	r, err = http.NewRequest(http.MethodGet, "/to/path", nil)
	a.NotError(err).NotNil(r)
	ctx := context.WithValue(r.Context(), contextKeyParams, &syntax.Params{Params: kvs})
	r = r.WithContext(ctx)
	ps2 := GetParams(r).(*syntax.Params)
	a.Equal(ps2.Params, kvs)
}
