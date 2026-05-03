package middleware

import (
	"github.com/charviki/maze-cradle/httputil"
)

// CORS 返回 CORS 中间件（允许所有来源），委托给 cradle/httputil。
// 保留 middleware 包作为 Hertz 中间件统一入口，调用者无需关心底层实现包。
var CORS = httputil.CORS

// CORSWithOrigins 返回基于允许来源列表的 CORS 中间件，委托给 cradle/httputil。
var CORSWithOrigins = httputil.CORSWithOrigins
