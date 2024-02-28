// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package options 构建路由的参数
package options

import (
	"net/http"

	"github.com/issue9/mux/v7/internal/syntax"
)

type (
	Option func(*Options)

	Options struct {
		Lock         bool
		CORS         *CORS
		Interceptors *syntax.Interceptors
		URLDomain    string
		RecoverFunc  RecoverFunc
	}

	RecoverFunc func(http.ResponseWriter, any)
)

func Build(o ...Option) (*Options, error) {
	ret := &Options{Interceptors: syntax.NewInterceptors()}
	for _, opt := range o {
		opt(ret)
	}

	if err := ret.sanitize(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (o *Options) sanitize() error {
	if o.CORS == nil {
		o.CORS = &CORS{}
	}
	if err := o.CORS.sanitize(); err != nil {
		return err
	}

	l := len(o.URLDomain)
	if l != 0 && o.URLDomain[l-1] == '/' {
		o.URLDomain = o.URLDomain[:l-1]
	}

	return nil
}
