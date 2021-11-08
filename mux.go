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

// Option 自定义路由参数的函数原型
type Option = options.Option

// InterceptorFunc 拦截器的函数原型
type InterceptorFunc = syntax.InterceptorFunc

// Interceptor 针对带参数类型路由的拦截处理
//
// 在解析诸如 /authors/{id:\\d+} 带参数的路由项时，
// 用户可以通过拦截并自定义对参数部分 {id:\\d+} 的解析，
// 从而不需要走正则表达式的那一套解析流程，可以在一定程度上增强性能。
//
// 一旦正则表达式被拦截，则节点类型也将不再是正则表达式，
// 其处理优先级会比正则表达式类型高。 在某些情况下，可能会造成处理结果不相同。比如：
//  /authors/{id:\\d+}     // 1
//  /authors/{id:[0-9]+}   // 2
// 以上两条记录是相同的，但因为表达式不同，也能正常添加，
// 处理流程，会按添加顺序优先比对第一条，所以第二条是永远无法匹配的。
// 但是如果你此时添加了 Interceptor(InterceptorDigit, "[0-9]+")，
// 使第二个记录的优先级提升，会使第一条永远无法匹配到数据。
func Interceptor(f InterceptorFunc, name ...string) Option {
	return func(o *options.Options) {
		if o.Interceptors == nil {
			o.Interceptors = syntax.NewInterceptors()
		}
		o.Interceptors.Add(f, name...)
	}
}

// InterceptorAny 任意非空字符的拦截器
func InterceptorAny(path string) bool { return syntax.MatchAny(path) }

// InterceptorDigit 任意数字字符的拦截器
func InterceptorDigit(path string) bool { return syntax.MatchDigit(path) }

// InterceptorWord 任意英文单词的拦截器
func InterceptorWord(path string) bool { return syntax.MatchWord(path) }

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
//
// NOTE: AllowedCORS 与 CORS 属于相同功能的修改，会相互覆盖。
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
//
// NOTE: AllowedCORS 与 CORS 属于相同功能的修改，会相互覆盖。
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

var syntaxCheckerInterceptors = syntax.NewInterceptors()

// CheckSyntax 检测路由项的语法格式
//
// 路由中可通过 {} 指定参数名称，如果参数名中带 :，则 : 之后的为参数的约束条件，
// 比如 /posts/{id}.html 表示匹配任意任意字符的参数 id。/posts/{id:\d+}.html，
// 表示匹配正则表达式 \d+ 的参数 id。；
func CheckSyntax(pattern string) error {
	_, err := syntaxCheckerInterceptors.Split(pattern)
	return err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}
