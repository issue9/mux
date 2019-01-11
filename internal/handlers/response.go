// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import "net/http"

// 用于 head 请求的返回，过滤掉输出内容。
type response struct {
	http.ResponseWriter
}

func (resp *response) Write(data []byte) (int, error) {
	return 0, nil
}
