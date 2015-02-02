// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// 错误状态处理函数。
//
// msg详细的错误信息；code错误状态码。
type ErrorHandler func(w http.ResponseWriter, msg string, code int)

// 默认的ErrorHandler实现，直接调用http.Error()实现。
func defaultErrorHandler(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}
