// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

var _ Matcher = &Method{}

func TestMethod(t *testing.T) {
	defHandler := func(w http.ResponseWriter, r *http.Request) bool {
		return true
	}

	defFunc := MatcherFunc(defHandler)

	fn := func(method string, m Matcher, wont bool) {
		r, err := http.NewRequest(method, "", nil)
		assert.NotError(t, err)
		assert.Equal(t, m.ServeHTTP2(nil, r), wont)
	}

	m := NewMethod().Get(defFunc)
	fn("GET", m, true)
	fn("POST", m, false)

	m = NewMethod().Get(defFunc).Post(defFunc)
	fn("POST", m, true)
	fn("GET", m, true)
	fn("OPTIONS", m, false)
}
