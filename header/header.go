// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

//go:generate go run ./generate_problems.go

// Package header 定义了 HTTP 请求中的报头
package header

const (
	DNT             = "DNT"
	ProxyConnection = "Proxy-Connection"

	XRequestedWith         = "X-Requested-With"
	XForwardedFor          = "X-Forwarded-For"
	XRealIP                = "X-Real-IP"
	XRequestID             = "X-Request-ID"
	XPoweredBy             = "X-Powered-By"
	XUACompatible          = "X-UA-Compatible"
	XContentDuration       = "X-Content-Duration"
	XDNSPrefetchControl    = "X-DNS-Prefetch-Control"
	XPingback              = "X-Pingback"
	XRobotsTag             = "X-Robots-Tag"
	XDownloadOptions       = "X-Download-Options"
	XRateLimitLimit        = "X-Rate-Limit-Limit"
	XRateLimitRemaining    = "X-Rate-Limit-Remaining"
	XRateLimitReset        = "X-Rate-Limit-Reset"
	XForwardedHost         = "X-Forwarded-Host"
	XForwardedProto        = "X-Forwarded-Proto"
	XForwardedProtocol     = "X-Forwarded-Protocol"
	XForwardedSSL          = "X-Forwarded-Ssl"
	XUrlScheme             = "X-Url-Scheme"
	FrontEndHttps          = "Front-End-Https"
	XHttpMethodOverride    = "X-Http-Method-Override"
	XATTDeviceId           = "X-ATT-Deviceid"
	XWapProfile            = "X-Wap-Profile"
	XCsrfToken             = "X-Csrf-Token"
	XXSSProtection         = "X-XSS-Protection"
	XContentSecurityPolicy = "X-Content-Security-Policy"
	XWebKitCSP             = "X-WebKit-CSP"
)
