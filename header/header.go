// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package header 定义了 HTTP 请求中的报头
package header

// HTTP 中的报头名称
//
// 大部分内容源自 https://zh.wikipedia.org/wiki/HTTP%E5%A4%B4%E5%AD%97%E6%AE%B5
const (
	Accept                             = "Accept"          // 能够接受的回应内容类型（Content-Types）。	Accept: text/plain
	AcceptCharset                      = "Accept-Charset"  // 能够接受的字符集	Accept-Charset: utf-8
	AcceptEncoding                     = "Accept-Encoding" // 能够接受的编码方式列表。	Accept-Encoding: gzip, deflate
	AcceptLanguage                     = "Accept-Language" // 能够接受的回应内容的自然语言列表。	Accept-Language: en-US
	AcceptDatetime                     = "Accept-Datetime" // 能够接受的按照时间来表示的版本	Accept-Datetime: Thu, 31 May 2007 20:35:00 GMT	临时
	Authorization                      = "Authorization"   // 用于超文本传输协议的认证的认证信息	Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	CacheControl                       = "Cache-Control"   // 用来指定在这次的请求/响应链中的所有缓存机制 都必须遵守的指令	Cache-Control: no-cache
	Connection                         = "Connection"      // 该浏览器想要优先使用的连接类型	Connection: keep-alive Connection: Upgrade
	Cookie                             = "Cookie"          // 之前由服务器通过 Set-Cookie （下文详述）发送的一个超文本传输协议Cookie	Cookie: $Version=1; Skin=new;
	ContentLength                      = "Content-Length"  // 以八位字节数组 （8位的字节）表示的请求体的长度	Content-Length: 348
	ContentMD5                         = "Content-MD5"     // 请求体的内容的二进制 MD5 散列值，以 Base64 编码的结果	Content-MD5: Q2hlY2sgSW50ZWdyaXR5IQ==	过时的
	ContentDigest                      = "Content-Digest"
	WantContentDigest                  = "Want-Content-Digest"
	ReprDigest                         = "Repr-Digest"
	WantReprDigest                     = "Want-Repr-Digest"
	ContentType                        = "Content-Type"                // 请求体的 MIME 类型 （用于 POST 和 PUT 请求中）	Content-Type: application/x-www-form-urlencoded
	Date                               = "Date"                        // 发送该消息的日期和时间(按照 RFC7231 中定义的"超文本传输协议日期"格式来发送)	Date: Tue, 15 Nov 1994 08:12:31 GMT
	Expect                             = "Expect"                      // 表明客户端要求服务器做出特定的行为	Expect: 100-continue
	From                               = "From"                        // 发起此请求的用户的邮件地址	From: user@example.com
	Host                               = "Host"                        // 服务器的域名(用于虚拟主机)，以及服务器所监听的传输控制协议端口号。如果所请求的端口是对应的服务的标准端口，则端口号可被省略。自超文件传输协议版本1.1（HTTP/1.1）开始便是必需字段。	Host: zh.wikipedia.org:80 Host: zh.wikipedia.org
	IfMatch                            = "If-Match"                    // 仅当客户端提供的实体与服务器上对应的实体相匹配时，才进行对应的操作。主要作用时，用作像 PUT 这样的方法中，仅当从用户上次更新某个资源以来，该资源未被修改的情况下，才更新该资源。	If-Match: "737060cd8c284d8af7ad3082f209582d"
	IfModifiedSince                    = "If-Modified-Since"           // 允许在对应的内容未被修改的情况下返回 304 未修改（ 304 Not Modified ）	If-Modified-Since: Sat, 29 Oct 1994 19:43:31 GMT
	IfNoneMatch                        = "If-None-Match"               // 允许在对应的内容未被修改的情况下返回 304 未修改（ 304 Not Modified ），参考 超文本传输协议 的实体标记	If-None-Match: "737060cd8c284d8af7ad3082f209582d"
	IfRange                            = "If-Range"                    // 如果该实体未被修改过，则向我发送我所缺少的那一个或多个部分；否则，发送整个新的实体	If-Range: "737060cd8c284d8af7ad3082f209582d"
	IfUnmodifiedSince                  = "If-Unmodified-Since"         // 仅当该实体自某个特定时间已来未被修改的情况下，才发送回应。	If-Unmodified-Since: Sat, 29 Oct 1994 19:43:31 GMT
	MaxForwards                        = "Max-Forwards"                // 限制该消息可被代理及网关转发的次数。	Max-Forwards: 10
	Origin                             = "Origin"                      // 发起一个针对跨域源资源共享的请求（要求服务器在回应中加入一个‘访问控制-允许来源’（'Access-Control-Allow-Origin'）字段）。	Origin: http://www.example-social-network.com
	Pragma                             = "Pragma"                      // 与具体的实现相关，这些字段可能在请求/回应链中的任何时候产生多种效果。	Pragma: no-cache	常设但不常用
	ProxyAuthorization                 = "Proxy-Authorization"         // 用来向代理进行认证的认证信息。	Proxy-Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==
	Range                              = "Range"                       // 仅请求某个实体的一部分。字节偏移以0开始。	Range: bytes=500-999
	Referer                            = "Referer"                     // 表示浏览器所访问的前一个页面，正是那个页面上的某个链接将浏览器带到了当前所请求的这个页面。	Referer: http://zh.wikipedia.org/wiki/Main_Page
	TE                                 = "TE"                          // 浏览器预期接受的传输编码方式：可使用回应协议头 Transfer-Encoding 字段中的值；另外还可用"trailers"（与"分块 "传输方式相关）这个值来表明浏览器希望在最后一个尺寸为0的块之后还接收到一些额外的字段。	TE: trailers, deflate
	UserAgent                          = "User-Agent"                  // 浏览器的浏览器身份标识字符串	User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:12.0) Gecko/20100101 Firefox/21.0
	Upgrade                            = "Upgrade"                     // 要求服务器升级到另一个协议。	Upgrade: HTTP/2.0, SHTTP/1.3, IRC/6.9, RTA/x11
	Via                                = "Via"                         // 向服务器告知，这个请求是由哪些代理发出的。	Via: 1.0 fred, 1.1 example.com (Apache/1.1)
	Warning                            = "Warning"                     // 一个一般性的警告，告知，在实体内容体中可能存在错误。	Warning: 199 Miscellaneous warning
	AccessControlAllowOrigin           = "Access-Control-Allow-Origin" // 指定哪些网站可参与到跨来源资源共享过程中	Access-Control-Allow-Origin: *	临时
	AccessControlRequestMethod         = "Access-Control-Request-Method"
	AccessControlMaxAge                = "Access-Control-Max-Age"
	AccessControlAllowHeaders          = "Access-Control-Allow-Headers"
	AccessControlAllowMethods          = "Access-Control-Allow-Methods"
	AccessControlRequestHeaders        = "Access-Control-Request-Headers"
	AccessControlExposeHeaders         = "Access-Control-Expose-Headers"
	AccessControlAllowCredentials      = "Access-Control-Allow-Credentials"
	AccessControlAllowPrivateNetwork   = "Access-Control-Allow-Private-Network"
	AccessControlRequestPrivateNetwork = "Access-Control-Request-Private-Network"
	AcceptPost                         = "Accept-Post"
	AcceptPatch                        = "Accept-Patch"              // 指定服务器支持的文件格式类型。	Accept-Patch: text/example;charset=utf-8
	AcceptRanges                       = "Accept-Ranges"             // 这个服务器支持哪些种类的部分内容范围	Accept-Ranges: bytes
	Age                                = "Age"                       // 这个对象在代理缓存中存在的时间，以秒为单位	Age: 12
	Allow                              = "Allow"                     // 对于特定资源有效的动作。针对 HTTP/405 这一错误代码而使用	Allow: GET, HEAD
	ContentDisposition                 = "Content-Disposition"       // 一个可以让客户端下载文件并建议文件名的头部。文件名需要用双引号包裹。	Content-Disposition: attachment; filename="fname.ext"
	ContentEncoding                    = "Content-Encoding"          // 在数据上使用的编码类型。参考 超文本传输协议压缩 。	Content-Encoding: gzip
	ContentLanguage                    = "Content-Language"          // 内容所使用的语言	Content-Language: da
	ContentLocation                    = "Content-Location"          // 所返回的数据的一个候选位置	Content-Location: /index.htm
	ContentRange                       = "Content-Range"             // 这条部分消息是属于某条完整消息的哪个部分	Content-Range: bytes 21010-47021/47022
	ETag                               = "ETag"                      // 对于某个资源的某个特定版本的一个标识符，通常是一个 消息散列	ETag: "737060cd8c284d8af7ad3082f209582d"
	Expires                            = "Expires"                   // 指定一个日期/时间，超过该时间则认为此回应已经过期	Expires: Thu, 01 Dec 1994 16:00:00 GMT
	LastModified                       = "Last-Modified"             // 所请求的对象的最后修改日期(按照 RFC 7231 中定义的“超文本传输协议日期”格式来表示)	Last-Modified: Tue, 15 Nov 1994 12:45:26 GMT
	Link                               = "Link"                      // 用来表达与另一个资源之间的类型关系，此处所说的类型关系是在 RFC 5988 中定义的	Link: </feed>; rel="alternate"[31]
	Location                           = "Location"                  // 用来进行重定向，或者在创建了某个新资源时使用。	Location: http://www.w3.org/pub/WWW/People.html
	P3P                                = "P3P"                       // 用于支持设置P3P策略，标准格式为“P3P:CP="your_compact_policy"”。然而P3P规范并不成功，[32]大部分现代浏览器没有完整实现该功能，而大量网站也将该值设为假值，从而足以用来欺骗浏览器的P3P插件功能并授权给第三方Cookies。	P3P: CP="This is not a P3P policy! See http://www.google.com/support/accounts/bin/answer.py?hl=en&answer=151657 for more info."
	ProxyAuthenticate                  = "Proxy-Authenticate"        // 要求在访问代理时提供身份认证信息。	Proxy-Authenticate: Basic
	PublicKeyPins                      = "Public-Key-Pins"           // 用于缓解中间人攻击，声明网站认证使用的传输层安全协议证书的散列值	Public-Key-Pins: max-age=2592000; pin-sha256="E9CZ9INDbd+2eRQozYqqbQ2yXLVKB9+xcprMF+44U1g=";
	Refresh                            = "Refresh"                   // 用于设定可定时的重定向跳转。右边例子设定了5秒后跳转至“http://www.w3.org/pub/WWW/People.html”。	Refresh: 5; url=http://www.w3.org/pub/WWW/People.html。
	RetryAfter                         = "Retry-After"               // 如果某个实体临时不可用，则，此协议头用来告知客户端日后重试。其值可以是一个特定的时间段(以秒为单位)或一个超文本传输协议日期。 	Example 1: Retry-After: 120 Example 2: Retry-After: Fri, 07 Nov 2014 23:59:59 GMT
	Server                             = "Server"                    // 服务器的名字	Server: Apache/2.4.1 (Unix)
	SetCookie                          = "Set-Cookie"                // HTTP cookie	Set-Cookie: UserID=JohnDoe; Max-Age=3600; Version=1
	Status                             = "Status"                    // 通用网关接口协议头字段，用来说明当前这个超文本传输协议回应的状态。普通的超文本传输协议回应，会使用单独的“状态行”（"Status-Line"）作为替代，这一点是在 RFC 7230 中定义的。 [35]	Status: 200 OK	Not listed as a registered field name（页面存档备份，存于互联网档案馆）
	StrictTransportSecurity            = "Strict-Transport-Security" // HTTP 严格传输安全这一头部告知客户端缓存这一强制 HTTPS 策略的时间，以及这一策略是否适用于其子域名。	Strict-Transport-Security: max-age=16070400; includeSubDomains
	Trailer                            = "Trailer"                   // 这个头部数值指示了在这一系列头部信息由由分块传输编码编码。	Trailer: Max-Forwards
	TransferEncoding                   = "Transfer-Encoding"         // 用来将实体安全地传输给用户的编码形式。当前定义的方法包括：分块（chunked）、compress、deflate、gzip和identity。	Transfer-Encoding: chunked
	Vary                               = "Vary"                      // 告知下游的代理服务器，应当如何对未来的请求协议头进行匹配，以决定是否可使用已缓存的回应内容而不是重新从原始服务器请求新的内容。	Vary:
	WWWAuthenticate                    = "WWW-Authenticate"          // 表明在请求获取这个实体时应当使用的认证模式。	WWW-Authenticate: Basic
	ClearSiteData                      = "Clear-Site-Data"
	AcceptCH                           = "Accept-CH"
	AcceptCHLifetime                   = "Accept-CH-Lifetime"
	ContentDPR                         = "Content-DPR"
	DPR                                = "DPR"
	EarlyData                          = "Early-Data"
	SaveData                           = "Save-Data"
	ViewportWidth                      = "Viewport-Width"
	Width                              = "Width"
	KeepAlive                          = "Keep-Alive"
	TimingAllowOrigin                  = "Timing-Allow-Origin"
	Tk                                 = "Tk"
	Forwarded                          = "Forwarded"
	ContentSecurityPolicyReportOnly    = "Content-Security-Policy-Report-Only"
	CrossOriginResourcePolicy          = "Cross-Origin-Resource-Policy"
	PublicKeyPinsReportOnly            = "Public-Key-Pins-Report-Only"
	UpgradeInsecureRequests            = "Upgrade-Insecure-Requests"
	LastEventID                        = "Last-Event-ID"
	NEL                                = "NEL"
	PingFrom                           = "Ping-From"
	PingTo                             = "Ping-To"
	ReportTo                           = "Report-To"
	SecWebSocketAccept                 = "Sec-WebSocket-Accept"
	SecWebSocketExtensions             = "Sec-WebSocket-Extensions"
	SecWebSocketKey                    = "Sec-WebSocket-Key"
	SecWebSocketProtocol               = "Sec-WebSocket-Protocol"
	SecWebSocketVersion                = "Sec-WebSocket-Version"
	AcceptPushPolicy                   = "Accept-Push-Policy"
	AcceptSignature                    = "Accept-Signature"
	AltSvc                             = "Alt-Svc"
	Index                              = "Index"
	LargeAllocation                    = "Large-Allocation"
	PushPolicy                         = "Push-Policy"
	ServerTiming                       = "Server-Timing"
	Signature                          = "Signature"
	SignedHeaders                      = "Signed-Headers"
	SourceMap                          = "SourceMap"

	// 点击劫持保护：
	//
	// deny：该页面不允许在 frame 中展示，即使是同域名内。
	// sameorigin：该页面允许同域名内在 frame 中展示。
	// allow-from uri：该页面允许在指定 uri 的 frame 中展示。
	// allowall：允许任意位置的frame显示，非标准值。
	// X-Frame-Options: deny	过时的
	XFrameOptions = "X-Frame-Options"

	// 非标准报头

	XRequestedWith      = "X-Requested-With"  // 主要用于标识 Ajax 及可扩展标记语言 请求。大部分的 JavaScript 框架会发送这个字段，且将其值设置为 XMLHttpRequest	X-Requested-With: XMLHttpRequest
	DNT                 = "DNT"               // 请求某个网页应用程序停止跟踪某个用户。在火狐浏览器中，相当于 X-Do-Not-Track 协议头字段（自 Firefox/4.0 Beta 11 版开始支持）。Safari 和 Internet Explorer 9 也支持这个字段。2011年3月7日，草案提交IETF。 万维网协会的跟踪保护工作组正在就此制作一项规范。	DNT: 1 (DNT启用)	DNT: 0 (DNT被禁用)
	XForwardedFor       = "X-Forwarded-For"   // 一个事实标准 ，用于标识某个通过超文本传输协议代理或负载均衡连接到某个网页服务器的客户端的原始互联网地址	X-Forwarded-For: client1, proxy1, proxy2 X-Forwarded-For: 129.78.138.66, 129.78.64.103
	XForwardedHost      = "X-Forwarded-Host"  // 一个事实标准 ，用于识别客户端原本发出的 Host 请求头部。	X-Forwarded-Host: zh.wikipedia.org:80 X-Forwarded-Host: zh.wikipedia.org
	XForwardedProto     = "X-Forwarded-Proto" // 一个事实标准，用于标识某个超文本传输协议请求最初所使用的协议。	X-Forwarded-Proto: https
	XRealIP             = "X-Real-IP"
	XForwardedProtocol  = "X-Forwarded-Protocol"
	XForwardedSSL       = "X-Forwarded-Ssl"
	XUrlScheme          = "X-Url-Scheme"
	FrontEndHttps       = "Front-End-Https"        // 被微软的服务器和负载均衡器所使用的非标准头部字段。	Front-End-Https: on
	XHttpMethodOverride = "X-Http-Method-Override" // 请求某个网页应用程序使用该协议头字段中指定的方法（一般是PUT或DELETE）来覆盖掉在请求中所指定的方法（一般是 POST）。当某个浏览器或防火墙阻止直接发送PUT 或DELETE 方法时（注意，这可能是因为软件中的某个漏洞，因而需要修复，也可能是因为某个配置选项就是如此要求的，因而不应当设法绕过），可使用这种方式。	X-HTTP-Method-Override: DELETE
	XATTDeviceId        = "X-ATT-Deviceid"         // 使服务器更容易解读 AT&T 设备 User-Agent 字段中常见的设备型号、固件信息。	X-Att-Deviceid: GT-P7320/P7320XXLPG
	XWapProfile         = "X-Wap-Profile"          // 链接到互联网上的一个 XML 文件，其完整、仔细地描述了正在连接的设备。右侧以为 AT&T Samsung Galaxy S2 提供的 XML 文件为例。	x-wap-profile: http://wap.samsungmobile.com/uaprof/SGH-I777.xml
	ProxyConnection     = "Proxy-Connection"       // 该字段源于早期超文本传输协议版本实现中的错误。与标准的连接（Connection）字段的功能完全相同。	Proxy-Connection: keep-alive
	XCsrfToken          = "X-Csrf-Token"           // 用于防止 跨站请求伪造。 辅助用的头部有 X-CSRFToken 或 X-XSRF-TOKEN	X-Csrf-Token: i8XNjC4b8KVok4uw5RftR38Wgp2BFwql
	XXSSProtection      = "X-XSS-Protection"       // 跨站脚本攻击 （XSS）过滤器	X-XSS-Protection: 1; mode=block

	// 内容安全策略定义。	X-WebKit-CSP: default-src 'self'
	ContentSecurityPolicy  = "Content-Security-Policy"
	XContentSecurityPolicy = "X-Content-Security-Policy"
	XWebKitCSP             = "X-WebKit-CSP"

	XContentTypeOptions           = "X-Content-Type-Options"            // 唯一允许的数值为 "nosniff"，防止 Internet Explorer 对文件进行MIME类型嗅探。这也对 Google Chrome 下载扩展时适用。X-Content-Type-Options: nosniff
	XPoweredBy                    = "X-Powered-By"                      // 表明用于支持当前网页应用程序的技术（例如：PHP）（版本号细节通常放置在 X-Runtime 或 X-Version 中）	X-Powered-By: PHP/5.4.0
	XUACompatible                 = "X-UA-Compatible"                   // 推荐指定的渲染引擎（通常是向后兼容模式）来显示内容。也用于激活 Internet Explorer 中的 Chrome Frame。	X-UA-Compatible: IE=EmulateIE7	X-UA-Compatible: IE=edge X-UA-Compatible: Chrome=1
	XContentDuration              = "X-Content-Duration"                // 指出音视频的长度，单位为秒。只受 Gecko 内核浏览器支持。	X-Content-Duration: 42.666
	FeaturePolicy                 = "Feature-Policy"                    // 管控特定应用程序接口	Feature-Policy: vibrate 'none'; geolocation 'none'
	PermissionsPolicy             = "Permissions-Policy"                // 管控特定应用程序接口为 W3C 标准 替代 Feature-Policy	Permissions-Policy: microphone=(),geolocation=(),camera=()
	XPermittedCrossDomainPolicies = "X-Permitted-Cross-Domain-Policies" // Flash 的跨网站攻击防御	X-Permitted-Cross-Domain-Policies: none
	ReferrerPolicy                = "Referrer-Policy"                   // 保护信息泄漏	//Referrer-Policy: origin-when-cross-origin
	ExpectCT                      = "Expect-CT"                         // 防止欺骗 SSL，单位为秒	Expect-CT: max-age=31536000, enforce
	XDNSPrefetchControl           = "X-DNS-Prefetch-Control"
	XPingback                     = "X-Pingback"
	XRequestID                    = "X-Request-ID"
	XRobotsTag                    = "X-Robots-Tag"
	XDownloadOptions              = "X-Download-Options"
	XRateLimitLimit               = "X-Rate-Limit-Limit"
	XRateLimitRemaining           = "X-Rate-Limit-Remaining"
	XRateLimitReset               = "X-Rate-Limit-Reset"
)
