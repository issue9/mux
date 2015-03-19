// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"strconv"
)

// 带http状态的错误信息。
type StatusError struct {
	Code    int
	Message string
}

func (e *StatusError) Error() string {
	return strconv.Itoa(e.Code) + ":" + e.Message
}
