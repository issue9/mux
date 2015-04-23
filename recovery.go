// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

// 错误处理函数。
// msg为输出的错误信息，可能是任意类型的数据，一般为从recover()返回的数据。
type RecoverFunc func(w http.ResponseWriter, msg interface{})

// ErrorHandlerFunc的默认实现。
// 为一个简单的500错误信息。不会输出msg参数的内容。
func defaultRecoverFunc(w http.ResponseWriter, msg interface{}) {
	http.Error(w, http.StatusText(500), 500)
}

// 捕获并处理panic信息。
type recovery struct {
	handler     http.Handler
	recoverFunc RecoverFunc
}

// 声明一个错误处理的handler，h参数中发生的panic将被截获并处理，不会再向上级反映。
// 当h参数为空时，直接panic。
// rf参数用于指定处理panic信息的函数，其原型为RecoverFunc，
// 当将rf指定为nil时，将使用默认的处理函数，
// 仅仅向客户端输出500的错误信息，没有具体内容。
func NewRecovery(h http.Handler, rf RecoverFunc) *recovery {
	if h == nil {
		panic("NewRecovery:参数h不能为空")
	}

	if rf == nil {
		rf = defaultRecoverFunc
	}

	return &recovery{
		handler:     h,
		recoverFunc: rf,
	}
}

// implement net/http.Handler.ServeHTTP(...)
func (r *recovery) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			r.recoverFunc(w, err)
		}
	}()

	r.handler.ServeHTTP(w, req)
}
