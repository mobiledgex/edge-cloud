package orm

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mobiledgex/edge-cloud/log"
)

func logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		if strings.HasSuffix(req.RequestURI, "show") || strings.HasSuffix(req.RequestURI, "showall") {
			// don't log show commands
			return next(c)
		}

		reqBody := []byte{}
		resBody := []byte{}
		var nexterr error
		if !strings.HasSuffix(req.RequestURI, "login") {
			// use body dump to capture req/res.
			// This makes a copy of the req and resp so we
			// avoid this unless needed.
			bd := middleware.BodyDump(func(c echo.Context, reqB, resB []byte) {
				reqBody = reqB
				resBody = resB
			})
			handler := bd(next)
			nexterr = handler(c)
			if strings.HasSuffix(req.RequestURI, "usercreate") {
				// req has password so don't log it
				reqBody = []byte{}
			}
		} else {
			nexterr = next(c)
		}
		kvs := []interface{}{}
		kvs = append(kvs, "method")
		kvs = append(kvs, req.Method)
		kvs = append(kvs, "uri")
		kvs = append(kvs, req.RequestURI)
		kvs = append(kvs, "remote-ip")
		kvs = append(kvs, c.RealIP())
		if claims, err := getClaims(c); err == nil {
			kvs = append(kvs, "user")
			kvs = append(kvs, claims.Username)
		}
		kvs = append(kvs, "status")
		kvs = append(kvs, res.Status)
		if nexterr != nil {
			kvs = append(kvs, "err")
			kvs = append(kvs, nexterr)
		}
		if len(reqBody) > 0 {
			kvs = append(kvs, "req")
			kvs = append(kvs, string(reqBody))
		}
		if len(resBody) > 0 {
			kvs = append(kvs, "resp")
			kvs = append(kvs, string(resBody))
		}

		log.InfoLog("Audit", kvs...)
		return nexterr
	}
}
