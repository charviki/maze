package middleware

import (
	"github.com/charviki/maze-cradle/httputil"
)

// CORS 返回 CORS 中间件（允许所有来源），委托给 cradle/httputil
var CORS = httputil.CORS

// CORSWithOrigins 返回基于允许来源列表的 CORS 中间件
var CORSWithOrigins = httputil.CORSWithOrigins
