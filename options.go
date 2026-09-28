// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/issue9/source"

	"github.com/issue9/mux/v10/internal/cors"
	"github.com/issue9/mux/v10/internal/syntax"
)

type (
	Option func(*options)

	options struct {
		caseInsensitive bool
		trace           bool
		lock            bool
		cors            *cors.CORS
		interceptors    *syntax.Interceptors
		pathPrefix      string
		recoverFunc     RecoverFunc
	}

	RecoverFunc = func(http.ResponseWriter, any)

	InterceptorFunc = syntax.InterceptorFunc
)

// WithCaseInsensitive 是否不区分大小写
//
// 该行为只针对客户端的请求地址，会将 [Request.URL.Path] 转换为小写与现有的路由项进行对比，
// 但是不会改变 [Request.URL.Path] 本身。
//
// 该开关不会影响由 [Router.Add] 等一系列添加路由项的 pattern 参数，
// 如果这些 pattern 参数为大写，可能永远无法匹配任何地址。
func WithCaseInsensitive(v bool) Option { return func(o *options) { o.caseInsensitive = v } }

// WithTrace 是否启用 TRACE 方法
func WithTrace(v bool) Option { return func(o *options) { o.trace = v } }

// WithLock 是否加锁
//
// 在调用 [Router.Handle] 等添加路由时，有可能会改变整个路由树的结构，
// 如果需要频繁在运行时添加和删除路由项，那么应当添加此选项。
func WithLock(l bool) Option { return func(o *options) { o.lock = l } }

// WithPathPrefix 为 [Router.URL] 生成的地址添加前缀
func WithPathPrefix(prefix string) Option { return func(o *options) { o.pathPrefix = prefix } }

// WithRecovery 用于指定路由 panic 之后的处理方法
//
// 如果多次指定，则最后一次启作用。
func WithRecovery(f RecoverFunc) Option { return func(o *options) { o.recoverFunc = f } }

// WithStatusRecovery 仅向客户端输出 status 状态码
func WithStatusRecovery(status int) Option {
	return WithRecovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
	})
}

// WithWriteRecovery 向 [io.Writer] 输出错误信息
//
// status 表示向客户端输出的状态码；
// out 表示输出通道，比如 [os.Stderr] 等；
func WithWriteRecovery(status int, out io.Writer) Option {
	return WithRecovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		source.DumpStack(out, 4, true, msg)
	})
}

// WithLogRecovery 将错误信息输出到日志
//
// status 表示向客户端输出的状态码；
// l 为输出的日志；
func WithLogRecovery(status int, l *slog.Logger) Option {
	return WithRecovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		l.Error(source.Stack(4, true, msg))
	})
}

// WithInterceptor 针对带参数类型路由的拦截处理
//
// 在解析诸如 /authors/{id:\\d+} 带参数的路由项时，
// 用户可以通过拦截并自定义对参数部分 {id:\\d+} 的解析，
// 从而不需要走正则表达式的那一套解析流程，可以在一定程度上增强性能。
//
// 一旦正则表达式被拦截，则节点类型也将不再是正则表达式，
// 其处理优先级会比正则表达式类型高。 在某些情况下，可能会造成处理结果不相同。比如：
//
//	/authors/{id:\\d+}     // 1
//	/authors/{id:[0-9]+}   // 2
//
// 以上两条记录是相同的，但因为表达式不同，也能正常添加，
// 处理流程，会按添加顺序优先比对第一条，所以第二条是永远无法匹配的。
// 但是如果你此时添加了 (InterceptorDigit, "[0-9]+")，
// 使第二个记录的优先级提升，会使第一条永远无法匹配到数据。
//
// 可多次调用，表示同时指定了多个。
func WithInterceptor(f InterceptorFunc, rule ...string) Option {
	return func(o *options) { o.interceptors.Add(f, rule...) }
}

// WithAnyInterceptor 任意非空字符的拦截器
func WithAnyInterceptor(rule string) Option { return WithInterceptor(syntax.MatchAny, rule) }

// WithDigitInterceptor 任意数字字符的拦截器
func WithDigitInterceptor(rule string) Option { return WithInterceptor(syntax.MatchDigit, rule) }

// WithWordInterceptor 任意英文单词的拦截器
func WithWordInterceptor(rule string) Option { return WithInterceptor(syntax.MatchWord, rule) }

// WithCORS 自定义[跨域请求]设置项
//
// origin 对应 Access-Control-Allow-Origin 报头。如果包含了 *，那么其它的设置将不再启作用。
// 如果此值为空，表示不启用跨域的相关设置；
//
// allowHeaders 对应 Access-Control-Allow-Headers
// 可以包含 *，表示可以是任意值，其它值将不再启作用；
//
// exposedHeaders 对应 Access-Control-Expose-Headers；
//
// maxAge 对应 Access-Control-Max-Age 有以下几种取值：
//   - 0 不输出该报头；
//   - -1 表示禁用；
//   - 其它 >= -1 的值正常输出数值；
//
// allowCredentials 对应 Access-Control-Allow-Credentials；
//
// [跨域请求]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors
func WithCORS(origin []string, allowHeaders []string, exposedHeaders []string, maxAge int, allowCredentials bool) Option {
	return func(o *options) {
		o.cors = &cors.CORS{
			Origins:          origin,
			AllowHeaders:     allowHeaders,
			ExposedHeaders:   exposedHeaders,
			MaxAge:           maxAge,
			AllowCredentials: allowCredentials,
		}
	}
}

// WithDenyCORS 禁用跨域请求的 [WithCORS] 选项
func WithDenyCORS() Option { return WithCORS(nil, nil, nil, 0, false) }

// WithAllowedCORS 允许跨域请求的 [WithCOORS] 选项
func WithAllowedCORS(maxAge int) Option {
	return WithCORS([]string{"*"}, []string{"*"}, nil, maxAge, false)
}

func buildOption(o ...Option) (*options, error) {
	ret := &options{interceptors: syntax.NewInterceptors()}
	for _, opt := range o {
		opt(ret)
	}

	if err := ret.sanitize(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (o *options) sanitize() error {
	if o.cors == nil {
		o.cors = &cors.CORS{}
	}
	if err := o.cors.Sanitize(); err != nil {
		return err
	}

	l := len(o.pathPrefix)
	if l != 0 && o.pathPrefix[l-1] == '/' {
		o.pathPrefix = o.pathPrefix[:l-1]
	}

	return nil
}
