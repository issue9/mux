// SPDX-License-Identifier: MIT

package mux

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/issue9/mux/v6/internal/syntax"
	"github.com/issue9/mux/v6/internal/tree"
)

type Option func(*options)

type RecoverFunc func(http.ResponseWriter, any)

type options struct {
	CaseInsensitive bool
	Lock            bool
	CORS            *cors
	Interceptors    *syntax.Interceptors
	URLDomain       string
	RecoverFunc     RecoverFunc

	NotFound,
	MethodNotAllowed http.Handler
}

type cors struct {
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

// Recovery 用于指路由 panic 之后的处理方法
//
// 除了采用 Option 的方式之外，也可以使用中间件的方式处理 panic，
// 但是中间件的处理方式需要用户保证其在中间件的最外层，
// 否则可能会发生其外层的中间件发生 panic 无法捕获的问题。
// 而 Option 方式没有此问题。
func Recovery(f RecoverFunc) Option {
	return func(o *options) { o.RecoverFunc = f }
}

// HTTPRecovery 仅向客户端输出 status 状态码
func HTTPRecovery(status int) Option {
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

// URLDomain 为 RouterOf.URL 生成的地址带上域名
func URLDomain(domain string) Option {
	return func(o *options) { o.URLDomain = domain }
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
	return func(o *options) {
		if o.Interceptors == nil {
			o.Interceptors = syntax.NewInterceptors()
		}
		o.Interceptors.Add(f, name...)
	}
}

// InterceptorAny 任意非空字符的拦截器
func InterceptorAny(rule string) bool { return syntax.MatchAny(rule) }

// InterceptorDigit 任意数字字符的拦截器
func InterceptorDigit(rule string) bool { return syntax.MatchDigit(rule) }

// InterceptorWord 任意英文单词的拦截器
func InterceptorWord(rule string) bool { return syntax.MatchWord(rule) }

// CaseInsensitive 忽略大小写
//
// 该操作仅是将客户端的请求路径转换为小之后再次进行匹配，
// 如果服务端的路由项设置为大写，则依然是不匹配的。
func CaseInsensitive(o *options) { o.CaseInsensitive = true }

// Lock 是否加锁
//
// 在调用 RouterOf.Add 添加路由时，有可能会改变整个路由树的结构，
// 如果需要频繁在运行时添加和删除路由项，那么应当添加此选项。
func Lock(o *options) { o.Lock = true }

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
	return func(o *options) {
		o.CORS = &cors{
			Origins:          origins,
			AllowHeaders:     allowHeaders,
			ExposedHeaders:   exposedHeaders,
			MaxAge:           maxAge,
			AllowCredentials: allowCredentials,
		}
	}
}

// AllowedCORS 允许跨域请求
func AllowedCORS(o *options) {
	o.CORS = &cors{
		Origins:      []string{"*"},
		AllowHeaders: []string{"*"},
		MaxAge:       3600,
	}
}

// DenyCORS 禁用跨域请求
func DenyCORS(o *options) { o.CORS = &cors{} }

// NotFound 自定义 404 状态码下的输出
func NotFound(h http.Handler) Option {
	return func(o *options) { o.NotFound = h }
}

// MethodNotAllowed 自定义 405 状态码下的输出
//
// 在 405 状态码下，除了输出用户指定的输出内容之外，系统还会输出 Allow 报头。
func MethodNotAllowed(h http.Handler) Option {
	return func(o *options) { o.MethodNotAllowed = h }
}

func (o *options) sanitize() error {
	if o.CORS == nil {
		o.CORS = &cors{}
	}
	if err := o.CORS.sanitize(); err != nil {
		return err
	}

	if o.Interceptors == nil {
		o.Interceptors = syntax.NewInterceptors()
	}

	l := len(o.URLDomain)
	if l != 0 && o.URLDomain[l-1] == '/' {
		o.URLDomain = o.URLDomain[:l-1]
	}

	if o.NotFound == nil {
		o.NotFound = http.NotFoundHandler()
	}

	if o.MethodNotAllowed == nil {
		o.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		})
	}

	return nil
}

func (o *options) handleCORS(node *tree.Node, w http.ResponseWriter, r *http.Request) {
	o.CORS.handle(node, w, r)
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

func (c *cors) handle(node *tree.Node, w http.ResponseWriter, r *http.Request) {
	if c.deny {
		return
	}

	// Origin 是可以为空的，所以采用 Access-Control-Request-Method 判断是否为预检。
	reqMethod := r.Header.Get("Access-Control-Request-Method")
	preflight := r.Method == http.MethodOptions &&
		reqMethod != "" &&
		r.URL.Path != "*" // OPTIONS * 不算预检，也不存在其它的请求方法处理方式。

	wh := w.Header()

	if preflight {
		// Access-Control-Allow-Methods
		methods := node.Methods()
		if !inStrings(methods, reqMethod) {
			return
		}
		wh.Set("Access-Control-Allow-Methods", node.Options())
		wh.Add("Vary", "Access-Control-Request-Method")

		// Access-Control-Allow-Headers
		if !c.headerIsAllowed(r) {
			return
		}
		if c.allowHeadersString != "" {
			wh.Set("Access-Control-Allow-Headers", c.allowHeadersString)
			wh.Add("Vary", "Access-Control-Request-Headers")
		}

		// Access-Control-Max-Age
		if c.maxAgeString != "" {
			wh.Set("Access-Control-Max-Age", c.maxAgeString)
		}
	}

	// Access-Control-Allow-Origin
	allowOrigin := "*"
	if !c.anyOrigins {
		origin := r.Header.Get("Origin")
		if !inStrings(c.Origins, origin) {
			return
		}
		allowOrigin = origin
	}
	wh.Set("Access-Control-Allow-Origin", allowOrigin)
	wh.Add("Vary", "Origin")

	// Access-Control-Allow-Credentials
	if c.AllowCredentials {
		wh.Set("Access-Control-Allow-Credentials", "true")
	}

	// Access-Control-Expose-Headers
	if c.exposedHeadersString != "" {
		wh.Set("Access-Control-Expose-Headers", c.exposedHeadersString)
	}
}

func (c *cors) headerIsAllowed(r *http.Request) bool {
	if c.anyHeaders {
		return true
	}

	h := strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers"))
	if h == "" {
		return true
	}

	headers := strings.Split(h, ",")
	for _, v := range headers {
		v = strings.TrimSpace(v)
		if !inStrings(c.AllowHeaders, v) {
			return false
		}
	}

	return true
}

func inStrings(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func buildOptions(o ...Option) (*options, error) {
	opt := &options{}
	for _, option := range o {
		if option == nil {
			panic("option 不能为空值")
		}

		option(opt)
	}

	if err := opt.sanitize(); err != nil {
		return nil, err
	}
	return opt, nil
}
