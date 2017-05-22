// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package list

import (
	"net/http"

	"github.com/issue9/assert"
	"github.com/issue9/mux/internal/syntax"
)

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	h1 = http.HandlerFunc(f1)
)

func newSyntax(a *assert.Assertion, pattern string) *syntax.Syntax {
	s, err := syntax.New(pattern)
	a.NotError(err).NotNil(s)
	return s
}
