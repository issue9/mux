// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package options

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/issue9/mux/v7/header"
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

func (c *CORS) headerIsAllowed(r *http.Request) bool {
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
