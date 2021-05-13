// SPDX-License-Identifier: MIT

package mux

// CORS 跨域请求设置项
//
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS
type CORS struct {
	// AllowedOrigins 允许的外部域名列表
	//
	// 可以是 *，如果包含了 *，那么其它的设置将不再启作用。
	// 此字段将被用于与请求头的 Origin 字段作验证，以确定是否放行该请求。
	AllowedOrigins []string

	// AllowedMethods 实际请求所允许使用的请求方法
	//
	// 可以是 *
	AllowedMethods []string

	// AllowedHeaders 实际请求中允许携带的报头
	//
	// 应该始终包含 Origin 报头，可以是 *。
	AllowedHeaders []string

	// ExposedHeaders Access-Control-Expose-Headers
	ExposedHeaders []string

	// MaxAge 当前报头信息可补缓存的秒数
	MaxAge int

	// AllowCredentials 是否允许 cookie
	AllowCredentials bool
}
