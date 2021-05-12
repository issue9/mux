// SPDX-License-Identifier: MIT

package handlers

import "net/http"

// 用于 head 请求的返回，过滤掉输出内容。
type response struct {
	http.ResponseWriter
}

func (resp *response) Write([]byte) (int, error) { return 0, nil }
