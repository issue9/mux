// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
package mux

import (
	"net/http"

	"github.com/issue9/mux/v5/internal/options"
	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/internal/tree"
	"github.com/issue9/mux/v5/params"
)

// Option 声明了自定义路由参数的函数原型
type Option func(options *options.Options)

// CaseInsensitive 忽略大小写
//
// 该操作仅是将客户端的请求路径转换为小之后再次进行匹配，
// 如果服务端的路由项设置为大写，则依然是不匹配的。
func CaseInsensitive(o *options.Options) { o.CaseInsensitive = true }

// Lock 是否加锁
//
// 在调用 Router.Add 添加路由时，有可能会改变整个路由树的结构，
// 如果需要频繁在运行时添加和删除路由项，那么应当添加此选项。
func Lock(o *options.Options) { o.Lock = true }

// CORS 自定义跨域请求设置项
//
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors
//
// origins 对应 Origin 报头。如果包含了 *，那么其它的设置将不再启作用。
// 如果此值为空，表示不启用跨域的相关设置；
//
// allowHeaders 对应 Access-Control-Allow-Headers
// 可以包含 *，表示可以是任意值，其它值将不再启作用；
//
// exposedHeaders 对应 Access-Control-Expose-Headers
//
// maxAge 对应 Access-Control-Max-Age 有以下几种取值：
// 0 不输出该报头；
// -1 表示禁用；
// 其它 >= -1 的值正常输出数值；
//
// allowCredentials 对应 Access-Control-Allow-Credentials。
func CORS(origins, allowHeaders, exposedHeaders []string, maxAge int, allowCredentials bool) Option {
	return func(o *options.Options) {
		o.CORS = &options.CORS{
			Origins:          origins,
			AllowHeaders:     allowHeaders,
			ExposedHeaders:   exposedHeaders,
			MaxAge:           maxAge,
			AllowCredentials: allowCredentials,
		}
	}
}

// AllowedCORS 允许跨域请求
func AllowedCORS(o *options.Options) { o.CORS = options.AllowedCORS() }

// NotFound 自定义 404 状态码下的输出
func NotFound(h http.Handler) Option {
	return func(o *options.Options) { o.NotFound = h }
}

// MethodNotAllowed 自定义 405 状态码下的输出
//
// 在 405 状态码下，除了输出用户指定的输出内容之外，系统还会输出 Allow 报头。
func MethodNotAllowed(h http.Handler) Option {
	return func(o *options.Options) { o.MethodNotAllowed = h }
}

// Params 获取路由中的参数集合
func Params(r *http.Request) params.Params { return params.Get(r) }

// CheckSyntax 检测路由项的语法格式
//
// 路由中可通过 {} 指定参数名称，如果参数名中带 :，则 : 之后的为参数的约束条件，
// 比如 /posts/{id}.html 表示匹配任意任意字符的参数 id。/posts/{id:\d+}.html，
// 表示匹配正则表达式 \d+ 的参数 id。；
func CheckSyntax(pattern string) error {
	_, err := syntax.Split(pattern)
	return err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}
