// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"runtime"
)

// 错误处理函数。
//
// msg为输出的错误信息，可能是任意类型的数据，一般为从recover()返回的数据。
type RecoverFunc func(w http.ResponseWriter, msg interface{})

// RecoverFunc的默认实现。
//
// 为一个简单的500错误信息。不会输出msg参数的内容。
func defaultRecoverFunc(w http.ResponseWriter, msg interface{}) {
	http.Error(w, http.StatusText(500), 500)
}

// RecoverFunc类型的实现。方便NewRecovery在调度期间将函数的调用信息输出到w。
func PrintDebug(w http.ResponseWriter, msg interface{}) {
	fmt.Fprintln(w, msg)
	for i := 1; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return
		}

		fmt.Fprintf(w, "@ %v:%v\n", file, line)
	}
}

// 捕获并处理panic信息。
type recovery struct {
	handler     http.Handler
	recoverFunc RecoverFunc
}

// 声明一个错误处理的handler，h参数中发生的panic将被截获并处理，不会再向上级反映。
//
// recovery应该处在所有http.Handler的最外层，用于处理所有没有被处理的panic。
//
// 当h参数为空时，将直接panic。
// rf参数用于指定处理panic信息的函数，其原型为RecoverFunc，
// 当将rf指定为nil时，将使用默认的处理函数，仅仅向客户端输出500的错误信息，没有具体内容。
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
