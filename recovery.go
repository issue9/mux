// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
)

// 错误处理函数。
// msg为输出的错误信息，可能是任意类型的数据。
type RecoverFunc func(w http.ResponseWriter, msg interface{})

// ErrorHandlerFunc的默认实现。
// msg为一个从recover()中返回的值。
func defaultRecoverFunc(w http.ResponseWriter, msg interface{}) {
	http.Error(w, fmt.Sprint(msg), 404)
}

type Recovery struct {
	handler     http.Handler
	recoverFunc RecoverFunc
}

// 声明一个错误处理的handler，h参数中发生的panic将被截获并处理，不会再向上级反映。
// 当h参数为空时，直接panic
func NewRecovery(h http.Handler, rf RecoverFunc) *Recovery {
	if h == nil {
		panic("NewRecovery:参数h不能为空")
	}

	if rf == nil {
		rf = defaultRecoverFunc
	}

	return &Recovery{
		handler:     h,
		recoverFunc: rf,
	}
}

func (r *Recovery) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			r.recoverFunc(w, err)
		}
	}()

	r.handler.ServeHTTP(w, req)
}
