// SPDX-License-Identifier: MIT

// Package muxutil 为 mux 提供的一些额外工具
package muxutil

import (
	"bytes"
	"html"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"path"
	"path/filepath"
)

const traceContentType = "message/http"

// Trace 简单的 Trace 请求方法实现
//
// NOTE: 并不是百分百原样返回，具体可参考 net/http/httputil.DumpRequest 的说明。
// 如果内容包含特殊的 HTML 字符会被 html.EscapeString 转码。
func Trace(w http.ResponseWriter, r *http.Request, body bool) error {
	text, err := httputil.DumpRequest(r, body)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", traceContentType)
		_, err = w.Write([]byte(html.EscapeString(string(text))))
	}

	return err
}

// ServeFile 提供对静态文件的服务
//
// p 表示需要读取的文件名；
// index 表示 p 为目录时，默认读取的文件，为空表示 index.html；
func ServeFile(fsys fs.FS, p, index string, w http.ResponseWriter, r *http.Request) error {
	if index == "" {
		index = "index.html"
	}

	if p == "" || p[len(p)-1] == '/' {
		p += index
	}

STAT:
	f, err := fsys.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		p = path.Join(p, index)
		goto STAT
	}

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		data := make([]byte, stat.Size())
		size, err := f.Read(data)
		if err != nil {
			return err
		}
		rs = bytes.NewReader(data[:size])
	}

	http.ServeContent(w, r, filepath.Base(p), stat.ModTime(), rs)
	return nil
}
