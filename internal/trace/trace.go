// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package trace

import (
	"html"
	"net/http"
	"net/http/httputil"

	"github.com/issue9/mux/v8/header"
)

// Trace 简单的 Trace 请求方法实现
//
// NOTE: 并不是百分百原样返回，具体可参考 [httputil.DumpRequest] 的说明。
// 如果内容包含特殊的 HTML 字符会被 [html.EscapeString] 转码。
func Trace(w http.ResponseWriter, r *http.Request, body bool) error {
	text, err := httputil.DumpRequest(r, body)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(header.ContentType, header.MessageHTTP)
		_, err = w.Write([]byte(html.EscapeString(string(text))))
	}

	return err
}
