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
	r = WithValue(r, &syntax.Params{Params: []syntax.Param{{K: "k1", V: "v1"}}})
	r = WithValue(r, &syntax.Params{Params: []syntax.Param{{K: "k2", V: "v2"}}})
	a.Equal(GetParams(r), &syntax.Params{Params: []syntax.Param{{K: "k2", V: "v2"}, {K: "k1", V: "v1"}}})
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
