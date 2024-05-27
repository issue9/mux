// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package header

// 一些报头的常用值
const (
	UTF8     = "utf-8"
	NoCache  = "no-cache"
	NoStore  = "no-store"
	Identity = "identity"
	Chunked  = "chunked"

	MessageHTTP       = "message/http"                      // TRACE 请求的 content-type 值
	MultipartFormData = "multipart/form-data"               // 表单提交类型
	FormData          = "application/x-www-form-urlencoded" // 普通的表单上传
	EventStream       = "text/event-stream"                 // SSE 的 content-type 值
	Plain             = "text/plain"
	HTML              = "text/html"
	JSON              = "application/json"
	XML               = "application/xml"
	Javascript        = "application/javascript"
	OctetStream       = "application/octet-stream"
)
