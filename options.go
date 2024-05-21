// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package mux

import (
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/issue9/source"

	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/internal/syntax"
	"github.com/issue9/mux/v8/types"
)

type (
	Option func(*options)

	options struct {
		trace        bool
		lock         bool
		cors         *cors
		interceptors *syntax.Interceptors
		urlDomain    string
		recoverFunc  RecoverFunc
	}

	cors struct {
		Origins    []string
		anyOrigins bool
		deny       bool

		AllowHeaders       []string
		allowHeadersString string
		anyHeaders         bool

		ExposedHeaders       []string
		exposedHeadersString string

		MaxAge       int
		maxAgeString string

		AllowCredentials bool
	}

	RecoverFunc = func(http.ResponseWriter, any)

	InterceptorFunc = syntax.InterceptorFunc
)

// Trace 是否启用 Trace 请求方法
func Trace(enable bool) Option { return func(o *options) { o.trace = enable } }

// Lock 是否加锁
//
// 在调用 [RouterOf.Handle] 等添加路由时，有可能会改变整个路由树的结构，
// 如果需要频繁在运行时添加和删除路由项，那么应当添加此选项。
func Lock(l bool) Option { return func(o *options) { o.lock = l } }

// URLDomain 为 [RouterOf.URL] 生成的地址带上域名
func URLDomain(prefix string) Option { return func(o *options) { o.urlDomain = prefix } }

// Recovery 用于指定路由 panic 之后的处理方法
//
// 如果多次指定，则最后一次启作用。
func Recovery(f RecoverFunc) Option { return func(o *options) { o.recoverFunc = f } }

// StatusRecovery 仅向客户端输出 status 状态码
func StatusRecovery(status int) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
	})
}

// WriteRecovery 向 [io.Writer] 输出错误信息
//
// status 表示向客户端输出的状态码；
// out 表示输出通道，比如 [os.Stderr] 等；
func WriteRecovery(status int, out io.Writer) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		source.DumpStack(out, 4, true, msg)
	})
}

// LogRecovery 将错误信息输出到日志
//
// status 表示向客户端输出的状态码；
// l 为输出的日志；
func LogRecovery(status int, l *log.Logger) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		l.Println(source.Stack(4, true, msg))
	})
}

// SLogRecovery 将错误信息输出到日志
//
// status 表示向客户端输出的状态码；
// l 为输出的日志；
func SLogRecovery(status int, l *slog.Logger) Option {
	return Recovery(func(w http.ResponseWriter, msg any) {
		http.Error(w, http.StatusText(status), status)
		l.Error(source.Stack(4, true, msg))
	})
}

// Interceptor 针对带参数类型路由的拦截处理
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
func Interceptor(f InterceptorFunc, rule ...string) Option {
	return func(o *options) { o.interceptors.Add(f, rule...) }
}

// AnyInterceptor 任意非空字符的拦截器
func AnyInterceptor(rule string) Option { return Interceptor(syntax.MatchAny, rule) }

// DigitInterceptor 任意数字字符的拦截器
func DigitInterceptor(rule string) Option { return Interceptor(syntax.MatchDigit, rule) }

// WordInterceptor 任意英文单词的拦截器
func WordInterceptor(rule string) Option { return Interceptor(syntax.MatchWord, rule) }

// CORS 自定义[跨域请求]设置项
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
//   - 0 不输出该报头；
//   - -1 表示禁用；
//   - 其它 >= -1 的值正常输出数值；
//
// allowCredentials 对应 Access-Control-Allow-Credentials；
//
// [跨域请求]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors
func CORS(origin []string, allowHeaders []string, exposedHeaders []string, maxAge int, allowCredentials bool) Option {
	return func(o *options) {
		o.cors = &cors{
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
		o.cors = &cors{}
	}
	if err := o.cors.sanitize(); err != nil {
		return err
	}

	l := len(o.urlDomain)
	if l != 0 && o.urlDomain[l-1] == '/' {
		o.urlDomain = o.urlDomain[:l-1]
	}

	return nil
}

func (c *cors) sanitize() error {
	for _, o := range c.Origins {
		if o == "*" {
			c.anyOrigins = true
			break
		}
	}
	c.deny = len(c.Origins) == 0

	for _, h := range c.AllowHeaders {
		if h == "*" {
			c.allowHeadersString = "*"
			c.anyHeaders = true
			break
		}
	}
	if c.allowHeadersString == "" && len(c.AllowHeaders) > 0 {
		c.allowHeadersString = strings.Join(c.AllowHeaders, ",")
	}

	if len(c.ExposedHeaders) > 0 {
		c.exposedHeadersString = strings.Join(c.ExposedHeaders, ",")
	}

	switch {
	case c.MaxAge == 0:
	case c.MaxAge >= -1:
		c.maxAgeString = strconv.Itoa(c.MaxAge)
	default:
		return errors.New("maxAge 的值只能是 >= -1")
	}

	if c.anyOrigins && c.AllowCredentials {
		return errors.New("origin=* 和 allowCredentials=true 不能同时成立")
	}

	return nil
}

func (c *cors) Handle(node types.Node, wh http.Header, r *http.Request) {
	if c.deny {
		return
	}

	// Origin 是可以为空的，所以采用 Access-Control-Request-Method 判断是否为预检。
	reqMethod := r.Header.Get(header.AccessControlRequestMethod)
	preflight := r.Method == http.MethodOptions &&
		reqMethod != "" &&
		r.URL.Path != "*" // OPTIONS * 不算预检，也不存在其它的请求方法处理方式。

	if preflight {
		// Access-Control-Allow-Methods
		if slices.Index(node.Methods(), reqMethod) < 0 {
			return
		}
		wh.Set(header.AccessControlAllowMethods, node.AllowHeader())
		wh.Add(header.Vary, header.AccessControlRequestMethod)

		// Access-Control-Allow-Headers
		if !c.headerIsAllowed(r) {
			return
		}
		if c.allowHeadersString != "" {
			wh.Set(header.AccessControlAllowHeaders, c.allowHeadersString)
			wh.Add(header.Vary, header.AccessControlAllowHeaders)
		}

		// Access-Control-Max-Age
		if c.maxAgeString != "" {
			wh.Set(header.AccessControlMaxAge, c.maxAgeString)
		}
	}

	// Access-Control-Allow-Origin
	allowOrigin := "*"
	if !c.anyOrigins {
		origin := r.Header.Get(header.Origin)
		if slices.Index(c.Origins, origin) < 0 {
			return
		}
		allowOrigin = origin
	}
	wh.Set(header.AccessControlAllowOrigin, allowOrigin)
	wh.Add(header.Vary, header.Origin)

	// Access-Control-Allow-Credentials
	if c.AllowCredentials {
		wh.Set(header.AccessControlAllowCredentials, "true")
	}

	// Access-Control-Expose-Headers
	if c.exposedHeadersString != "" {
		wh.Set(header.AccessControlExposeHeaders, c.exposedHeadersString)
	}
}

func (c *cors) headerIsAllowed(r *http.Request) bool {
	if c.anyHeaders {
		return true
	}

	h := strings.TrimSpace(r.Header.Get(header.AccessControlRequestHeaders))
	if h == "" {
		return true
	}

	headers := strings.Split(h, ",")
	for _, v := range headers {
		if slices.Index(c.AllowHeaders, strings.TrimSpace(v)) < 0 {
			return false
		}
	}

	return true
}
