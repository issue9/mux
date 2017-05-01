// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestResource_Clean(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	// 添加 delete /api/1
	a.NotPanic(func() {
		srvmux.DeleteFunc("/api/1", f1).
			PatchFunc("/api/1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	// 添加 patch /api/2 和 delete /api/2
	res, err := srvmux.Resource("/api/2")
	a.NotError(err).NotNil(res)
	a.NotPanic(func() {
		res.PatchFunc(f1).
			Delete(h1)
	})
	a.Equal(srvmux.entries.Len(), 2)

	res.Clean()
	a.Equal(srvmux.entries.Len(), 1)
}

func TestMux_Resource(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	res, err := srvmux.Resource("/abc/1")
	a.NotError(err).NotNil(res)
	a.Equal(res.Mux(), srvmux)
	a.Equal(res.pattern, "/abc/1")
}

func TestResource_URL(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	// 非正则
	res, err := srvmux.Resource("/api/v1")
	a.NotError(err).NotNil(res)
	url, err := res.URL(map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/v1")

	// 未命名正则
	res, err = srvmux.Resource("/api/{:\\d+}")
	a.NotError(err).NotNil(res)
	url, err = res.URL(map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/{:\\d+}")

	// 正常的单个参数
	res, err = srvmux.Resource("/api/{id:\\d+}")
	a.NotError(err).NotNil(res)
	url, err = res.URL(map[string]string{"id": "1"})
	a.NotError(err).Equal(url, "/api/1")

	// 多个参数
	res, err = srvmux.Resource("/api/{action}/{id:\\d+}")
	a.NotError(err).NotNil(res)
	url, err = res.URL(map[string]string{"id": "1", "action": "blog"})
	a.NotError(err).Equal(url, "/api/blog/1")
	// 缺少参数
	url, err = res.URL(map[string]string{"id": "1"})
	a.Error(err).Equal(url, "")
}
