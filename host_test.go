// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

// 错误处理函数。向response写入hostErrorHandler
func hostErrorHandler(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	w.Write([]byte("hostErrorHandler"))
}

// 默认的handler，向response写入host1Handler
func host1Handler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("host1Handler"))
}

func TestHost_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	// 正常匹配正则
	host := NewHost(hostErrorHandler)
	a.NotNil(host).
		Equal(ErrorHandler(hostErrorHandler), host.errorHandler)
	a.NotError(host.Add("?127.0.0.1:\\d+", http.HandlerFunc(host1Handler)))
	a.Equal(getHostResponse(a, host), "host1Handler")

	// 无法匹配的域名，缺少端口号
	host = NewHost(hostErrorHandler)
	a.NotNil(host)
	a.NotError(host.Add("127.0.0.1", http.HandlerFunc(host1Handler)))
	a.Equal(getHostResponse(a, host), "hostErrorHandler")
}

func getHostResponse(a *assert.Assertion, host *Host) []byte {
	// 创建测试服务器
	srv := httptest.NewServer(host)
	a.NotNil(srv)
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	a.NotError(err).NotNil(resp)

	p, err := ioutil.ReadAll(resp.Body)
	a.NotError(err).NotNil(p)

	return p
}
