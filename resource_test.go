// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMux_Resource(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	r1, err := srvmux.Resource("/abc/1")
	a.NotError(err).NotNil(r1)
	a.Equal(r1.Mux(), srvmux)
	a.Equal(r1.pattern, "/abc/1")

	r2, err := srvmux.Resource("/abc/1")
	a.NotError(err).NotNil(r2)
	a.False(r1 == r2)            // 不是同一个 *Resource
	a.True(r1.entry == r2.entry) // 但应该指向同一个 entry.Entry 实例
}

func TestResource_Name(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	res, err := srvmux.Resource("/posts/{id}")
	a.NotError(err).NotNil(res)
	a.NotError(res.Name("post"))
	// 应该是同一个
	a.Equal(srvmux.Name("post"), res)

	// 未指定名称，不存在
	a.Nil(srvmux.Name("author"))

	res, err = srvmux.Resource("/posts/{id}/author")
	a.NotError(err).NotNil(res)
	// 同名
	a.Error(res.Name("post"))
}

func TestResource_URL(t *testing.T) {
	a := assert.New(t)
	srvmux := New(false, false, nil, nil)
	a.NotNil(srvmux)

	// 非正则
	res, err := srvmux.Resource("/api/v1")
	a.NotError(err).NotNil(res)
	url, err := res.URL(map[string]string{"id": "1"}, "path")
	a.NotError(err).Equal(url, "/api/v1")

	// 未命名正则
	res, err = srvmux.Resource("/api/{:\\d+}")
	a.NotError(err).NotNil(res)
	url, err = res.URL(map[string]string{"id": "1"}, "path")
	a.NotError(err).Equal(url, "/api/[0-9]+")

	// 正常的单个参数
	res, err = srvmux.Resource("/api/{id:\\d+}/*")
	a.NotError(err).NotNil(res)
	url, err = res.URL(map[string]string{"id": "1"}, "path")
	a.NotError(err).Equal(url, "/api/1/path")

	// 多个参数
	res, err = srvmux.Resource("/api/{action}/{id:\\d+}")
	a.NotError(err).NotNil(res)
	url, err = res.URL(map[string]string{"id": "1", "action": "blog"}, "path")
	a.NotError(err).Equal(url, "/api/blog/1")
	// 缺少参数
	url, err = res.URL(map[string]string{"id": "1"}, "path")
	a.Error(err).Equal(url, "")
}
