// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

// package cors 跨域的相关操作
package cors

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/types"
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

func (c *CORS) Sanitize() error {
	if slices.Contains(c.Origins, "*") {
		c.anyOrigins = true
	}
	c.deny = len(c.Origins) == 0

	if slices.Contains(c.AllowHeaders, "*") {
		c.allowHeadersString = "*," + header.Authorization // Firefox 中 * 并不包含 Authorization 报头。
		c.anyHeaders = true
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
	reqMethod := r.Header.Get(header.AccessControlRequestMethod)
	preflight := r.Method == http.MethodOptions &&
		reqMethod != "" &&
		r.URL.Path != "*" // OPTIONS * 不算预检，也不存在其它的请求方法处理方式。

	if preflight {
		// Access-Control-Allow-Methods
		if !slices.Contains(node.Methods(), reqMethod) {
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
		if !slices.Contains(c.Origins, origin) {
			return
		}
		allowOrigin = origin
	}
	wh.Set(header.AccessControlAllowOrigin, allowOrigin)
	wh.Add(header.Vary, header.AccessControlAllowOrigin)

	// Access-Control-Allow-Credentials
	if c.AllowCredentials {
		wh.Set(header.AccessControlAllowCredentials, "true")
	}

	// Access-Control-Expose-Headers
	if c.exposedHeadersString != "" {
		wh.Set(header.AccessControlExposeHeaders, c.exposedHeadersString)
	}
}

func (c *CORS) headerIsAllowed(r *http.Request) bool {
	if c.anyHeaders {
		return true
	}

	if h := strings.TrimSpace(r.Header.Get(header.AccessControlRequestHeaders)); h != "" {
		for v := range strings.SplitSeq(h, ",") {
			if !slices.Contains(c.AllowHeaders, strings.TrimSpace(v)) {
				return false
			}
		}
	}

	return true
}
