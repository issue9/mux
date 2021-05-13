// SPDX-License-Identifier: MIT

package mux

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/issue9/sliceutil"

	"github.com/issue9/mux/v4/internal/handlers"
)

// CORS 跨域请求设置项
//
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS
type CORS struct {
	// AllowedOrigins 允许的外部域名列表
	//
	// 可以是 *，如果包含了 *，那么其它的设置将不再启作用。
	// 此字段将被用于与请求头的 Origin 字段作验证，以确定是否放行该请求。
	AllowedOrigins []string
	anyOrigins     bool

	// AllowedHeaders 实际请求中允许携带的报头
	//
	// 应该始终包含 Origin 报头，可以是 *。
	AllowedHeaders       []string
	allowedHeadersString string

	// ExposedHeaders Access-Control-Expose-Headers
	ExposedHeaders       []string
	exposedHeadersString string

	// MaxAge 当前报头信息可补缓存的秒数
	MaxAge       uint64
	maxAgeString string

	// AllowCredentials 是否允许 cookie
	AllowCredentials bool
}

func Allowed() *CORS {
	return &CORS{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		MaxAge:         3600,
	}
}

func Denied() *CORS {
	return &CORS{}
}

func (c *CORS) build() {
	for _, o := range c.AllowedOrigins {
		if o == "*" {
			c.anyOrigins = true
			break
		}
	}

	for _, h := range c.AllowedHeaders {
		if h == "*" {
			c.allowedHeadersString = "*"
			break
		}
	}
	if c.allowedHeadersString == "" && len(c.AllowedHeaders) > 0 {
		c.allowedHeadersString = strings.Join(c.AllowedHeaders, ", ")
	}

	if len(c.ExposedHeaders) > 0 {
		c.exposedHeadersString = strings.Join(c.ExposedHeaders, ", ")
	}

	c.maxAgeString = strconv.FormatUint(c.MaxAge, 10)
}

func (c *CORS) handle(hs *handlers.Handlers, w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	h := w.Header()

	// Access-Control-Allow-Origin
	allowOrigin := "*"
	if !c.anyOrigins {
		i := sliceutil.Index(c.AllowedOrigins, func(i int) bool { return c.AllowedOrigins[i] == origin })
		if i >= 0 {
			allowOrigin = origin
		}
	}
	h.Set("Access-Control-Allow-Origin", allowOrigin)

	// Access-Control-Allow-Methods
	h.Set("Access-Control-Allow-Methods", hs.Options())

	// Access-Control-Allow-Headers
	if c.allowedHeadersString != "" {
		h.Set("Access-Control-Allow-Headers", c.allowedHeadersString)
	}

	// Access-Control-Allow-Credentials
	if c.AllowCredentials {
		h.Set("Access-Control-Allow-Credentials", "true")
	}

	// Access-Control-Max-Age
	if c.maxAgeString != "" {
		h.Set("Access-Control-Max-Age", c.maxAgeString)
	}

	// Access-Control-Expose-Headers
	if c.exposedHeadersString != "" {
		h.Set("Access-Control-Expose-Headers", c.exposedHeadersString)
	}
}
