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

// 测试NewHost()和Host.Add()
func TestHost_New_Add(t *testing.T) {
	a := assert.New(t)

	// 默认的ErrorHandler为defaultErrorHandler
	h := NewHost(nil)
	a.NotNil(h).
		Equal(0, len(h.entries)).
		Equal(h.errorHandler, ErrorHandler(defaultErrorHandler))

	// 空的handler指针
	a.Error(h.Add("abc.example.com", nil))
	a.Equal(0, len(h.entries)).
		Equal(0, len(h.namedEntries))

	fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	a.NotError(h.Add("abc.example.com", fn))
	a.Equal(1, len(h.entries)).
		Equal(1, len(h.namedEntries))

	// 添加相同名称的
	a.Error(h.Add("abc.example.com", fn))
	a.Equal(1, len(h.entries)).
		Equal(1, len(h.namedEntries))

	// 添加不同的域名
	a.NotError(h.Add("?\\w+.example.com", fn))
	a.Equal(2, len(h.entries)).
		Equal(2, len(h.namedEntries))
}

func TestMethod_Add(t *testing.T) {
	a := assert.New(t)
	m := NewMethod(nil)
	a.NotNil(m)

}

func errorHandler1(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}
