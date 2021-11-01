// SPDX-License-Identifier: MIT

package mux

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v5/internal/tree"
)

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

// CORS 跨域请求设置项
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
	return func(r *Router) {
		r.cors = &cors{
			Origins:          origins,
			AllowHeaders:     allowHeaders,
			ExposedHeaders:   exposedHeaders,
			MaxAge:           maxAge,
			AllowCredentials: allowCredentials,
		}
	}
}

// AllowedCORS 允许跨域请求
func AllowedCORS(r *Router) {
	r.cors = &cors{
		Origins:      []string{"*"},
		AllowHeaders: []string{"*"},
		MaxAge:       3600,
	}
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
		return errors.New("MaxAge 的值只能是 >= -1")
	}

	if c.anyOrigins && c.AllowCredentials {
		return errors.New("AllowedOrigin=* 和 AllowCredentials=true 不能同时成立")
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
		if sliceutil.Index(methods, func(i int) bool { return methods[i] == reqMethod }) == -1 {
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
		i := sliceutil.Index(c.Origins, func(i int) bool { return c.Origins[i] == origin })
		if i == -1 {
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
		i := sliceutil.Index(c.AllowHeaders, func(i int) bool { return c.AllowHeaders[i] == v })
		if i == -1 {
			return false
		}
	}

	return true
}
