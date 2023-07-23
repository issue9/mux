// SPDX-License-Identifier: MIT

package options

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v7/types"
)

type CORS struct {
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

func (c *CORS) sanitize() error {
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

func (c *CORS) Handle(node types.Node, wh http.Header, r *http.Request) {
	if c.deny {
		return
	}

	// Origin 是可以为空的，所以采用 Access-Control-Request-Method 判断是否为预检。
	reqMethod := r.Header.Get("Access-Control-Request-Method")
	preflight := r.Method == http.MethodOptions &&
		reqMethod != "" &&
		r.URL.Path != "*" // OPTIONS * 不算预检，也不存在其它的请求方法处理方式。

	if preflight {
		// Access-Control-Allow-Methods
		methods := node.Methods()
		if !inStrings(methods, reqMethod) {
			return
		}
		wh.Set("Access-Control-Allow-Methods", node.AllowHeader())
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

func (c *CORS) headerIsAllowed(r *http.Request) bool {
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
	return sliceutil.Exists(strs, func(e string, _ int) bool { return e == s })
}
