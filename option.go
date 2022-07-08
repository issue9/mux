// SPDX-License-Identifier: MIT

package mux

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/issue9/mux/v7/internal/options"
	"github.com/issue9/mux/v7/internal/syntax"
)

type RecoverFunc = options.RecoverFunc

type Option = options.Option

// Lock 是否加锁
//
// 在调用 RouterOf.Add 添加路由时，有可能会改变整个路由树的结构，
// 如果需要频繁在运行时添加和删除路由项，那么应当添加此选项。
func Lock(l bool) Option { return func(o *options.Options) { o.Lock = l } }

// URLDomain 为 RouterOf.URL 生成的地址带上域名
func URLDomain(prefix string) Option {
	return func(o *options.Options) { o.URLDomain = prefix }
}

// Recover 用于指路由 panic 之后的处理方法
//
// 如果多次指定，则最后一次启作用。
func Recovery(f RecoverFunc) Option {
	return func(o *options.Options) { o.RecoverFunc = f }
}

// StatusRecovery 仅向客户端输出 status 状态码
func StatusRecovery(status int) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
	})
}

// WriterRecovery 向 io.Writer 输出错误信息
//
// status 表示向客户端输出的状态码；
// out 输出的 io.Writer，比如 os.Stderr 等；
func WriterRecovery(status int, out io.Writer) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		if _, err := fmt.Fprint(out, msg, "\n", string(debug.Stack())); err != nil {
			panic(err)
		}
	})
}

// LogRecovery 将错误信息输出到日志
//
// status 表示向客户端输出的状态码；
// l 为输出的日志；
func LogRecovery(status int, l *log.Logger) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		l.Println(msg, "\n", string(debug.Stack()))
	})
}

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
// 但是如果你此时添加了 (InterceptorDigit, "[0-9]+")，
// 使第二个记录的优先级提升，会使第一条永远无法匹配到数据。
//
// 可多次调用，表示同时指定了多个。
func Interceptor(f InterceptorFunc, rule ...string) Option {
	return func(o *options.Options) { o.Interceptors.Add(f, rule...) }
}

// AnyInterceptor 任意非空字符的拦截器
func AnyInterceptor(rule string) Option { return Interceptor(syntax.MatchAny, rule) }

// DigitInterceptor 任意数字字符的拦截器
func DigitInterceptor(rule string) Option { return Interceptor(syntax.MatchDigit, rule) }

// WordInterceptor 任意英文单词的拦截器
func WordInterceptor(rule string) Option { return Interceptor(syntax.MatchWord, rule) }

// CORS 自定义跨域请求设置项
//
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors
//
// origin 对应 Origin 报头。如果包含了 *，那么其它的设置将不再启作用。
// 如果此值为空，表示不启用跨域的相关设置；
//
// allowHeaders 对应 Access-Control-Allow-Headers
// 可以包含 *，表示可以是任意值，其它值将不再启作用；
//
// exposedHeaders 对应 Access-Control-Expose-Headers；
//
// maxAge 对应 Access-Control-Max-Age 有以下几种取值：
//  - 0 不输出该报头；
//  - -1 表示禁用；
//  - 其它 >= -1 的值正常输出数值；
//
//
// allowCredentials 对应 Access-Control-Allow-Credentials；
func CORS(origin []string, allowHeaders []string, exposedHeaders []string, maxAge int, allowCredentials bool) Option {
	return func(o *options.Options) {
		o.CORS = &options.CORS{
			Origins:          origin,
			AllowHeaders:     allowHeaders,
			ExposedHeaders:   exposedHeaders,
			MaxAge:           maxAge,
			AllowCredentials: allowCredentials,
		}
	}
}

// DenyCORS 禁用跨域请求
func DenyCORS() Option { return CORS(nil, nil, nil, 0, false) }

// AllowedCORS 允许跨域请求
func AllowedCORS(maxAge int) Option { return CORS([]string{"*"}, []string{"*"}, nil, maxAge, false) }
