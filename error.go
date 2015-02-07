// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"strconv"
)

// 带http状态的错误信息。
type StatusError struct {
	Code    int
	Message string
}

func NewStatusError(code int, msg string) error {
	return &StatusError{
		Code:    code,
		Message: msg,
	}
}

func (e *StatusError) Error() string {
	return strconv.Itoa(e.Code) + " " + e.Message
}

// 错误处理函数
type ErrorHandlerFunc func(http.ResponseWriter, interface{})

// ErrorHandlerFunc的默认实现。
// msg为一个从recover()中返回的值。
func defaultErrorHandlerFunc(w http.ResponseWriter, msg interface{}) {
	http.Error(w, fmt.Sprint(msg), 404)
}

type errorHandler struct {
	errFunc ErrorHandlerFunc
	handler http.Handler
}

// 声明一个错误处理的handler，h参数中发生的panic将被截获并处理，不会再向上级反映。
// 当h参数为空时，直接panic
func ErrorHandler(h http.Handler, fun ErrorHandlerFunc) *errorHandler {
	if h == nil {
		panic("参数h不能为空")
	}

	if fun == nil {
		fun = defaultErrorHandlerFunc
	}

	return &errorHandler{
		errFunc: fun,
		handler: h,
	}
}

func (e *errorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			e.errFunc(w, err)
		}
	}()

	e.handler.ServeHTTP(w, req)
}
