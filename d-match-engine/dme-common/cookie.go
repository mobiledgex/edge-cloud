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
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// temporarily hardcode the keys here, later we can get them via config file
// or perhaps from the controller
var dmePrivateKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

var dmePublicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA4ZgYPw7f7KO9zGsTia9p
gO47B0nn7MFxuUu8Lo30QHJLYprotY0IpNtjBxR7oGlUwwiFT7tMXKtsudIjx3Zf
jQxED4UUQN9bqtybpQbPCQNS5djaaESxdKSBCfUhfBt5aXajeqrnDifloe7a2Z4A
VZjy5jCdbEtbZVz2WH6Xus7CFUAJR++2/sJo/KPH0ZCJKd8B0JqW2BJr3vJWbSOV
5tp/cztbDvimP978Gc+IjcJJWibecvwPsFSD8MPwd5utGnxKTMX+4pweH+45PUFK
q3npINgniowHzlghIKBBjd1L/gh93U+2tBFIOcH1TPz9+e+eQgdz1wIh+C/swSSJ
sSpdi+gL8FVYAyRQnoJPIDHxaOYPfM8cyojhII0gs7xPxCghk8DoIm/ohfb0ntLs
Ypzzl1M8tNSJPb5Y2inQpplbUCSH2AgWRiGE7SUrMphoQ4+UTtLdGCTrLln1fv5Z
nc/fM5zpiuOs8FH8lMCFw/UHBuae+8sgXw0fhymDM+Kyw94Ghe5/IO2oltdoYozS
kfm9iBqc2n3xBYLrDJjgdLUCyXUh3tlO3bc5O/CXM5eTmp62PqoxNPIXkMP9ekh+
l+CiuBqG+a/qu6qJVvPH1zt5lN+z+VorKI0qM0bLkr9+2Qq01FPfaSUm9LfL3ggh
rsYPlqEPvtSrS1HPve+6akMCAwEAAQ==
-----END PUBLIC KEY-----`

type CookieKey struct {
	PeerIP  string `json:"peerip,omitempty"`
	DevName string `json:"devname,omitempty"`
	AppName string `json:"appname,omitempty"`
	AppVers string `json:"appvers,omitempty"`
}

type dmeClaims struct {
	jwt.StandardClaims
	Key *CookieKey `json:"key,omitempty"`
}

type ctxCookieKey struct{}

// returns Peer IP or Error
func VerifyCookie(cookie string) (*CookieKey, error) {

	if cookie == "" {
		log.WarnLog("missing cookie in VerifyCookie")

		return nil, fmt.Errorf("missing cookie")
	}
	claims := dmeClaims{}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(dmePublicKey))
	if err != nil {
		return nil, err
	}
	_, err = jwt.ParseWithClaims(cookie, &claims, func(token *jwt.Token) (verifykey interface{}, err error) {
		return pubKey, nil
	})
	if err != nil {
		log.WarnLog("error in verifycookie", "cookie", cookie, "err", err)
		return nil, err
	}
	if claims.Key == nil {
		log.InfoLog("no key parsed", "cookie", cookie, "err", err)
		return nil, errors.New("No Key data in cookie")
	}

	if claims.ExpiresAt < time.Now().Unix() {
		log.InfoLog("cookie is expired", "cookie", cookie, "expiresAt", claims.ExpiresAt)
		return nil, errors.New("Expired cookie")
	}

	log.DebugLog(log.DebugLevelDmereq, "verified cookie", "cookie", cookie, "expires", claims.ExpiresAt)
	return claims.Key, nil
}

func GenerateCookie(key *CookieKey, ctx context.Context) (string, error) {
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
			ExpiresAt: time.Now().AddDate(0, 0, 1).Unix(),
		},
		Key: key,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims)
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(dmePrivateKey))
	if err != nil {
		return "", err
	}
	cookie, err := tok.SignedString(signKey)
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
			return nil, err
		}
		ctx = NewCookieContext(ctx, ckey)
	}
	// call the handler
	return handler(ctx, req)
}

func NewCookieContext(ctx context.Context, ckey *CookieKey) context.Context {
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
