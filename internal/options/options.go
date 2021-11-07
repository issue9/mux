// SPDX-License-Identifier: MIT

// Package options 提供了初始化路由的参数对象
package options

import "net/http"

type Options struct {
	CaseInsensitive bool
	Lock            bool
	CORS            *CORS

	NotFound,
	MethodNotAllowed http.Handler
}

func (o *Options) Sanitize() error {
	if o.CORS == nil {
		o.CORS = &CORS{}
	}
	if err := o.CORS.sanitize(); err != nil {
		return err
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
