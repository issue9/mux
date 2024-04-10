// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package header

// 一些报头的常用值
const (
	UTF8    = "utf-8"
	NoCache = "no-cache"

	// MessageHTTP TRACE 请求方法的 content-type 值
	MessageHTTP = "message/http"

	// MultipartFormData 表单提交类型
	MultipartFormData = "multipart/form-data"

	// FormData 普通的表单上传
	FormData = "application/x-www-form-urlencoded"

	EventStream = "text/event-stream"
	Plain       = "text/plain"
	HTML        = "text/html"
	JSON        = "application/json"
	XML         = "application/xml"
	Javascript  = "application/javascript"
	OctetStream = "application/octet-stream"
)
