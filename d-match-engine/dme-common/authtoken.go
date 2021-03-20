package dmecommon

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mobiledgex/edge-cloud/log"
)

type authClaims struct {
	jwt.StandardClaims
	OrgName string `json:"orgname,omitempty"`
	AppName string `json:"appname,omitempty"`
	AppVers string `json:"appvers,omitempty"`
}

// VerifyAuthToken verifies the token against the provided public key.  JWT contents for devname,
// appname and appvers must match the contents of the token
func VerifyAuthToken(ctx context.Context, token string, pubkey string, devname string, appname string, appvers string) error {
	if token == "" {
		return fmt.Errorf("empty token")
	}

	authClaims := authClaims{}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubkey))
	if err != nil {
		return errors.New("unable to parse JWT token")
	}
	_, err = jwt.ParseWithClaims(token, &authClaims, func(token *jwt.Token) (verifykey interface{}, err error) {
		return pubKey, nil
	})

	if err != nil {
		log.InfoLog("error in parse claims", "token", token, "err", err)
		return err
	}

	//check that the values in the token match
	if devname != authClaims.OrgName {
		return errors.New("token organization mismatch")
	}
	if appname != authClaims.AppName {
		return errors.New("token appname mismatch")
	}
	if appvers != authClaims.AppVers {
		return errors.New("token appvers mismatch")
	}

	log.SpanLog(ctx, log.DebugLevelDmereq, "verified token", "token", token, "expires", authClaims.ExpiresAt)
	return nil
}

// GenerateAuthToken is used only for test purposes, as the DME never
// generates auth tokens it only verifies them
func GenerateAuthToken(privKeyFile string, appOrg string, appname string, appvers string, expireTime int64) (string, error) {
	privkey, err := ioutil.ReadFile(privKeyFile)
	if err != nil {
		return "", fmt.Errorf("Cannot read private key file %s -- %v", privKeyFile, err)
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256,
		authClaims{
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: expireTime},
			OrgName: appOrg,
			AppName: appname,
			AppVers: appvers,
		})
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privkey))
	if err != nil {
		return "", err
	}
	return tok.SignedString(signKey)
}
