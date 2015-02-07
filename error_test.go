// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestStatusError(t *testing.T) {
	err := NewStatusError(404, "not found")
	assert.Equal(t, err.Error(), "404 not found")
}

func TestDefaultErrorHandlerFunc(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	a.NotNil(w)

	defaultErrorHandlerFunc(w, "not found")
	a.Equal("not found\n", w.Body.String())
}

func TestErrorHandler(t *testing.T) {
	a := assert.New(t)

	// h参数传递空值
	a.Panic(func() {
		ErrorHandler(nil, nil)
	})

	// 指定fun参数为nil，能正确设置其值
	eh := ErrorHandler(&Host{}, nil)
	a.NotNil(eh.errFunc)
}
