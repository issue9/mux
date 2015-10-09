// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ RecoverFunc = defaultRecoverFunc
var _ RecoverFunc = PrintDebug

func TestDefaultRecoverFunc(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	a.NotNil(w)

	defaultRecoverFunc(w, "not found")
	a.Equal(http.StatusText(404)+"\n", w.Body.String())
}

func TestNewRecovery(t *testing.T) {
	a := assert.New(t)

	// h参数传递空值
	a.Panic(func() {
		NewRecovery(nil, nil)
	})

	// 指定fun参数为nil，能正确设置其值
	r := NewRecovery(&ServeMux{}, nil)
	a.NotNil(r.recoverFunc)
}

func TestPrintDebug(t *testing.T) {
	w := httptest.NewRecorder()
	PrintDebug(w, "PrintDebug")
	t.Log(w.Body.String())
}
