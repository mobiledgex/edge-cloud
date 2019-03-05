package orm

import (
	"encoding/json"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
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
		// use body dump to capture req/res.
		bd := middleware.BodyDump(func(c echo.Context, reqB, resB []byte) {
			reqBody = reqB
			resBody = resB
		})
		handler := bd(next)
		nexterr = handler(c)
		// remove passwords from requests so they aren't logged
		if strings.HasSuffix(req.RequestURI, "login") {
			login := ormapi.UserLogin{}
			err := json.Unmarshal(reqBody, &login)
			if err == nil {
				login.Password = ""
				reqBody, err = json.Marshal(login)
			}
			if err != nil {
				reqBody = []byte{}
			}
		}
		if strings.HasSuffix(req.RequestURI, "usercreate") {
			user := ormapi.User{}
			err := json.Unmarshal(reqBody, &user)
			if err == nil {
				user.Passhash = ""
				reqBody, err = json.Marshal(user)
			}
			if err != nil {
				reqBody = []byte{}
			}
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
