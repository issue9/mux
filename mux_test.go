// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	//"net/http/httputil"
	"testing"

	"github.com/issue9/assert"
)

func T1estHost_Add(t *testing.T) {
	a := assert.New(t)

	h := NewHost(nil)
	a.NotNil(h).Equal(0, len(h.entries))

	a.NotError(h.Add("abc.example.com", nil))
	a.Equal(1, len(h.entries))

	// 添加相同名称的
	a.Error(h.Add("abc.example.com", nil))
	a.Equal(1, len(h.entries))

	// 添加不同的域名
	a.NotError(h.Add("?\\w+.example.com", nil))
	a.Equal(2, len(h.entries))
}

func T1estMethod_Add(t *testing.T) {
	a := assert.New(t)
	m := NewMethod(nil)
	a.NotNil(m)

	// 未指定method，panic
	a.Panic(func() {
		m.Add("test", nil)
	})
}

func errorHandler1(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}
