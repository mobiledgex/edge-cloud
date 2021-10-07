package vault

import (
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/stretchr/testify/require"
)

func TestJwk(t *testing.T) {
	vault := process.Vault{
		Common: process.Common{
			Name: "vault",
		},
		DmeSecret:  "123456",
		Regions:    "local",
		ListenAddr: "http://127.0.0.1:8300",
	}
	vroles, err := vault.StartLocalRoles()
	require.Nil(t, err, "start local vault")
	defer vault.StopLocal()

	roles := vroles.RegionRoles["local"]

	// this represents a dme process accessing vault
	vaultAddr := vault.ListenAddr
	dmeAuth := NewAppRoleAuth(roles.DmeRoleID, roles.DmeSecretID)
	dmeConfig := NewConfig(vaultAddr, dmeAuth)
	jwks := JWKS{}
	jwks.Init(dmeConfig, "local", "dme")
	// vault local process puts two secrets to start
	err = jwks.updateKeys()
	require.Nil(t, err, "update keys")
	require.Equal(t, 2, len(jwks.Keys))
	require.Equal(t, 2, jwks.Meta.CurrentVersion)
	jwk, found := jwks.Keys[jwks.Meta.CurrentVersion]
	require.True(t, found)
	require.Equal(t, vault.DmeSecret, jwk.Secret, "jwt secret")

	claims := &TestClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().AddDate(0, 0, 1).Unix(),
		},
	}
	cookie, err := jwks.GenerateCookie(claims)
	require.Nil(t, err, "generate cookie")
	kid, _ := claims.GetKid()
	require.Equal(t, 2, kid)
	_, err = jwks.VerifyCookie(cookie, claims)
	require.Nil(t, err, "verify cookie")

	// put a new secret to rotate secrets
	rotatorAuth := NewAppRoleAuth(roles.RotatorRoleID, roles.RotatorSecretID)
	rotatorConfig := NewConfig(vaultAddr, rotatorAuth)
	newSecret := "abcdefg"
	err = PutSecret(rotatorConfig, "local", "dme", newSecret, "1m")
	require.Nil(t, err, "put secret")
	// simulate another dme that has the new key set
	jwks2 := JWKS{}
	jwks2.Init(dmeConfig, "local", "dme")
	err = jwks2.updateKeys()
	require.Nil(t, err, "update keys2")
	require.Equal(t, 3, jwks2.Meta.CurrentVersion)
	jwk, found = jwks2.Keys[jwks2.Meta.CurrentVersion]
	require.True(t, found)
	require.Equal(t, newSecret, jwk.Secret)

	// make sure newer key set can verify old cookie
	_, err = jwks2.VerifyCookie(cookie, claims)
	require.Nil(t, err, "verify cookie")

	// make sure older key set can be updated to verify new cookie
	jwks.lastUpdateAttempt = time.Now().Add(-time.Minute)
	cookie, err = jwks2.GenerateCookie(claims)
	require.Nil(t, err, "generate cookie")
	kid, _ = claims.GetKid()
	require.Equal(t, 3, kid)
	_, err = jwks.VerifyCookie(cookie, claims)
	require.Nil(t, err, "verify cookie")
	require.Equal(t, 3, jwks.Meta.CurrentVersion)
}

type TestClaims struct {
	jwt.StandardClaims
	Kid int
}

func (s *TestClaims) GetKid() (int, error) { return s.Kid, nil }
func (s *TestClaims) SetKid(kid int)       { s.Kid = kid }
