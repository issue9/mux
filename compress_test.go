// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
)

var _ http.ResponseWriter = &compressWriter{}
var _ http.Hijacker = &compressWriter{}
