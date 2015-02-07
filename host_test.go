// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

// 测试NewHost()和Host.Handle()
func TestHost_New_Handle(t *testing.T) {
	a := assert.New(t)

	// 默认的ErrorHandler为defaultErrorHandler
	h := NewHost()
	a.NotNil(h).Equal(0, len(h.entries))

	// 空的handler指针
	a.Error(h.Add("abc.example.com", nil))
	a.Equal(0, len(h.entries)).
		Equal(0, len(h.namedEntries))

	fn := func(w http.ResponseWriter, req *http.Request) {}

	a.NotError(h.AddFunc("abc.example.com", fn))
	a.Equal(1, len(h.entries)).
		Equal(1, len(h.namedEntries))

	// 添加相同名称的
	a.Error(h.AddFunc("abc.example.com", fn))
	a.Equal(1, len(h.entries)).
		Equal(1, len(h.namedEntries))

	// 添加不同的域名
	a.NotError(h.AddFunc("?\\w+.example.com", fn))
	a.Equal(2, len(h.entries)).
		Equal(2, len(h.namedEntries))
}

func TestHost_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	newHost := func(pattern string) http.Handler {
		h := NewHost()
		a.NotError(h.AddFunc(pattern, defaultHandler))
		return h
	}

	tests := []*handlerTester{
		&handlerTester{
			name:       "host正则匹配",
			h:          newHost("?127.0.0.1:(?P<port>\\d+)"),
			response:   "OK",
			statusCode: 200,
		},
		&handlerTester{
			name:       "host字符串不匹配",
			h:          newHost("127.0.0.1"),
			response:   "error",
			statusCode: 404,
		},
	}

	runHandlerTester(a, tests)
}
