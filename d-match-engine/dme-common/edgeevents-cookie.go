package dmecommon

import (
	"context"
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mobiledgex/edge-cloud/log"
)

type EdgeEventsCookieKey struct {
	ClusterOrg   string `json:"clusterorg,omitempty"`
	ClusterName  string `json:"clustername,omitempty"`
	CloudletOrg  string `json:"cloudletorg,omitempty"`
	CloudletName string `json:"cloudletname,omitempty"`
	Kid          int    `json:"kid,omitempty"`
}

type edgeEventsClaims struct {
	jwt.StandardClaims
	Key *EdgeEventsCookieKey `json:"key,omitempty"`
}

type ctxEdgeEventsCookieKey struct{}

func (e *edgeEventsClaims) GetKid() (int, error) {
	if e.Key == nil {
		return 0, fmt.Errorf("Invalid cookie, no key")
	}
	return e.Key.Kid, nil
}

func (e *edgeEventsClaims) SetKid(kid int) {
	e.Key.Kid = kid
}

func VerifyEdgeEventsCookie(ctx context.Context, cookie string) (*EdgeEventsCookieKey, error) {
	claims := edgeEventsClaims{}
	token, err := Jwks.VerifyCookie(cookie, &claims)
	if err != nil {
		log.InfoLog("error in verifyedgeeventscookie", "cookie", cookie, "err", err)
		return nil, err
	}
	if claims.Key == nil || !verifyEdgeEventsCookieKey(claims.Key) {
		log.InfoLog("no key parsed", "eecookie", cookie, "err", err)
		return nil, errors.New("No Key data in cookie")
	}
	if !token.Valid {
		log.InfoLog("edgeevents cookie is invalid or expired", "eecookie", cookie, "claims", claims)
		return nil, errors.New("invalid or expired cookie")
	}
	log.SpanLog(ctx, log.DebugLevelDmereq, "verified edgeevents cookie", "eecookie", cookie, "expires", claims.ExpiresAt)
	return claims.Key, nil
}

func verifyEdgeEventsCookieKey(key *EdgeEventsCookieKey) bool {
	if key.ClusterOrg == "" && key.ClusterName == "" && key.CloudletOrg == "" && key.CloudletName == "" {
		return false
	}
	return true
}

func GenerateEdgeEventsCookie(key *EdgeEventsCookieKey, ctx context.Context, cookieExpiration *time.Duration) (string, error) {
	claims := edgeEventsClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(*cookieExpiration).Unix(),
		},
		Key: key,
	}

	cookie, err := Jwks.GenerateCookie(&claims)
	log.SpanLog(ctx, log.DebugLevelDmereq, "generated edge events cookie", "key", key, "eecookie", cookie, "err", err)
	return cookie, err
}

func NewEdgeEventsCookieContext(ctx context.Context, eekey *EdgeEventsCookieKey) context.Context {
	if eekey == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxEdgeEventsCookieKey{}, eekey)
}

func EdgeEventsCookieFromContext(ctx context.Context) (eekey *EdgeEventsCookieKey, ok bool) {
	eekey, ok = ctx.Value(ctxEdgeEventsCookieKey{}).(*EdgeEventsCookieKey)
	return
}

func IsTheSameCluster(key1 *EdgeEventsCookieKey, key2 *EdgeEventsCookieKey) bool {
	return key1.CloudletOrg == key2.CloudletOrg && key1.CloudletName == key2.CloudletName && key1.ClusterOrg == key2.ClusterOrg && key2.ClusterName == key2.ClusterName
}
