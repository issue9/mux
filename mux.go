// SPDX-License-Identifier: MIT

// Package mux 功能完备的路由中间件
package mux

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/issue9/errwrap"

	"github.com/issue9/mux/v5/internal/options"
	"github.com/issue9/mux/v5/internal/syntax"
	"github.com/issue9/mux/v5/internal/tree"
	"github.com/issue9/mux/v5/params"
)

type (
	// Option 自定义路由参数的函数原型
	Option = options.Option

	// RecoverFunc 路由对 panic 的处理函数原型
	RecoverFunc = options.RecoverFunc

	// InterceptorFunc 拦截器的函数原型
	InterceptorFunc = syntax.InterceptorFunc

	// Params 路由参数
	Params = params.Params
)

// Recovery 用于指路由 panic 之后的处理方法
//
// 除了采用 Option 的方式之外，也可以使用中间件的方式处理 panic，
// 但是中间件的处理方式需要用户保证其在中间件的最外层，
// 否则可能会发生其外层的中间件发生 panic 无法捕获的问题。
// 而 Option 方式没有此问题。
func Recovery(f RecoverFunc) Option {
	return func(o *options.Options) { o.RecoverFunc = f }
}

// HTTPRecovery 仅向客户端输出 status 状态码
func HTTPRecovery(status int) Option {
	return Recovery(func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
	})
}

// WriterRecovery 向 io.Writer 输出错误信息
//
// status 表示向客户端输出的状态码；
// out 输出的 io.Writer，比如 os.Stderr 等；
func WriterRecovery(status int, out io.Writer) Option {
	return Recovery(func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
		if _, err := fmt.Fprint(out, msg); err != nil {
			panic(err)
		}
	})
}

// LogRecovery 将错误信息输出到日志
//
// status 表示向客户端输出的状态码；
// l 为输出的日志；
func LogRecovery(status int, l *log.Logger) Option {
	return Recovery(func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
		l.Println(msg)
	})
}

// URLDomain 为 Router.URL 生成的地址带上域名
func URLDomain(domain string) Option {
	return func(o *options.Options) { o.URLDomain = domain }
}

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
// NOTE: AllowedCORS 与 CORS 会相互覆盖。
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
// NOTE: AllowedCORS 与 CORS 会相互覆盖。
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

// GetParams 获取路由中的参数集合
func GetParams(r *http.Request) Params { return syntax.GetParams(r) }

var emptyInterceptors = syntax.NewInterceptors()

// CheckSyntax 检测路由项的语法格式
//
// 路由中可通过 {} 指定参数名称，如果参数名中带 :，则 : 之后的为参数的约束条件，
// 比如 /posts/{id}.html 表示匹配任意任意字符的参数 id。/posts/{id:\d+}.html，
// 表示匹配正则表达式 \d+ 的参数 id。
func CheckSyntax(pattern string) error {
	_, err := emptyInterceptors.Split(pattern)
	return err
}

// URL 根据参数生成地址
//
// pattern 为路由项的定义内容；
// params 为路由项中的参数，键名为参数名，键值为参数值。
//
// NOTE: 仅仅是将 params 填入到 pattern 中， 不会判断参数格式是否正确。
func URL(pattern string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return pattern, nil
	}

	buf := errwrap.StringBuilder{}
	buf.Grow(len(pattern))
	if err := emptyInterceptors.URL(&buf, pattern, params); err != nil {
		return "", err
	}
	return buf.String(), buf.Err
}

// Methods 返回所有支持的请求方法
func Methods() []string {
	methods := make([]string, len(tree.Methods))
	copy(methods, tree.Methods)
	return methods
}
