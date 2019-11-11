package dmecommon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/vault"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
)

var Jwks vault.JWKS

func InitVault(addr, region string) error {
	config, err := vault.BestConfig(addr)
	if err != nil {
		return err
	}
	Jwks.Init(config, region, "dme")
	Jwks.GoUpdate(nil)
	return nil
}

type CookieKey struct {
	PeerIP  string `json:"peerip,omitempty"`
	DevName string `json:"devname,omitempty"`
	AppName string `json:"appname,omitempty"`
	AppVers string `json:"appvers,omitempty"`
	Kid     int    `json:"kid,omitempty"`
}

type dmeClaims struct {
	jwt.StandardClaims
	Key *CookieKey `json:"key,omitempty"`
}

type ctxCookieKey struct{}

func (d *dmeClaims) GetKid() (int, error) {
	if d.Key == nil {
		return 0, fmt.Errorf("Invalid cookie, no key")
	}
	return d.Key.Kid, nil
}

func (d *dmeClaims) SetKid(kid int) {
	d.Key.Kid = kid
}

// returns Peer IP or Error
func VerifyCookie(cookie string) (*CookieKey, error) {

	claims := dmeClaims{}
	token, err := Jwks.VerifyCookie(cookie, &claims)
	if err != nil {
		log.InfoLog("error in verifycookie", "cookie", cookie, "err", err)
		return nil, err
	}
	if claims.Key == nil {
		log.InfoLog("no key parsed", "cookie", cookie, "err", err)
		return nil, errors.New("No Key data in cookie")
	}
	if !token.Valid {
		log.InfoLog("cookie is invalid or expired", "cookie", cookie, "claims", claims)
		return nil, errors.New("invalid or expired cookie")
	}
	// It is possible that the App was deleted after this cookie was issued.  If so, the cookie is no longer valid
	// Note that returning an error from from here will result in an Unauthorized error code, rather than a NotFound
	if !AppExists(claims.Key.DevName, claims.Key.AppName, claims.Key.AppVers) {
		return nil, fmt.Errorf("app not found -- developer: %s, app: %s, appvers: %s", claims.Key.DevName, claims.Key.AppName, claims.Key.AppVers)
	}
	log.DebugLog(log.DebugLevelDmereq, "verified cookie", "cookie", cookie, "expires", claims.ExpiresAt)
	return claims.Key, nil
}

func GenerateCookie(key *CookieKey, ctx context.Context, cookieExpiration *time.Duration) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("unable to get peer IP info")
	}

	// TODO:
	// This needs to validate that the Dev/App data sent by client
	// is in our database, and is not spoofed by the client.

	ss := strings.Split(p.Addr.String(), ":")
	if len(ss) != 2 {
		return "", errors.New("unable to parse peer address " + p.Addr.String())
	}
	key.PeerIP = ss[0]
	claims := dmeClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
			// 1 day expiration for now
			ExpiresAt: time.Now().Add(*cookieExpiration).Unix(),
		},
		Key: key,
	}

	cookie, err := Jwks.GenerateCookie(&claims)
	log.DebugLog(log.DebugLevelDmereq, "generated cookie", "key", key, "cookie", cookie, "err", err)
	return cookie, err
}

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	allow := false
	var cookie string

	switch typ := req.(type) {
	case *dme.RegisterClientRequest:
		// allow any
		allow = true
	case *dme.FindCloudletRequest:
		cookie = typ.SessionCookie
	case *dme.VerifyLocationRequest:
		cookie = typ.SessionCookie
	case *dme.GetLocationRequest:
		cookie = typ.SessionCookie
	case *dme.DynamicLocGroupRequest:
		cookie = typ.SessionCookie
	case *dme.AppInstListRequest:
		cookie = typ.SessionCookie
	case *dme.FqdnListRequest:
		cookie = typ.SessionCookie
	}
	if !allow {
		// Verify session cookie, add decoded CookieKey to context
		ckey, err := VerifyCookie(cookie)
		if err != nil {
			return nil, grpc.Errorf(codes.Unauthenticated, err.Error())
		}
		ctx = NewCookieContext(ctx, ckey)
	}
	// call the handler
	return handler(ctx, req)
}

func NewCookieContext(ctx context.Context, ckey *CookieKey) context.Context {
	if ckey == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxCookieKey{}, ckey)
}

func CookieFromContext(ctx context.Context) (ckey *CookieKey, ok bool) {
	ckey, ok = ctx.Value(ctxCookieKey{}).(*CookieKey)
	return
}

// PeerContext is a helper function to a context with peer info
func PeerContext(ctx context.Context, ip string, port int) context.Context {
	addr := net.TCPAddr{}
	addr.IP = net.ParseIP(ip)
	addr.Port = port
	pr := peer.Peer{Addr: &addr}
	return peer.NewContext(ctx, &pr)
}
