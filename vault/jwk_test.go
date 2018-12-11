package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestJwk(t *testing.T) {
	dir, err := os.Getwd()
	require.Nil(t, err, "getwd")
	rolesfile := dir + "/roles.yaml"
	fmt.Printf("roles file is %s\n", rolesfile)

	vault := process.Vault{
		Name:        "vault",
		DmeSecret:   "123456",
		McormSecret: "987664",
	}
	err = vault.Start("", process.WithRolesFile(rolesfile))
	require.Nil(t, err, "start vault")
	defer vault.Stop()
	defer os.Remove(rolesfile)

	// rolesfile contains the roleIDs/secretIDs needed to access vault
	dat, err := ioutil.ReadFile(rolesfile)
	require.Nil(t, err, "read rolesfile")
	roles := process.VaultRoles{}
	err = yaml.Unmarshal(dat, &roles)
	require.Nil(t, err, "unmarshal roles")

	// this represents a dme process accessing vault
	jwks := JWKS{}
	jwks.Init(process.VaultAddress, "dme", roles.DmeRoleID, roles.DmeSecretID)
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
	require.Equal(t, 2, claims.GetKid())
	_, err = jwks.VerifyCookie(cookie, claims)
	require.Nil(t, err, "verify cookie")

	// put a new secret to rotate secrets
	newSecret := "abcdefg"
	err = PutSecret(process.VaultAddress, roles.RotatorRoleID, roles.RotatorSecretID, "dme", newSecret, "1m")
	require.Nil(t, err, "put secret")
	// simulate another dme that has the new key set
	jwks2 := JWKS{}
	jwks2.Init(process.VaultAddress, "dme", roles.DmeRoleID, roles.DmeSecretID)
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
	require.Equal(t, 3, claims.GetKid())
	_, err = jwks.VerifyCookie(cookie, claims)
	require.Nil(t, err, "verify cookie")
	require.Equal(t, 3, jwks.Meta.CurrentVersion)
}

type TestClaims struct {
	jwt.StandardClaims
	Kid int
}

func (s *TestClaims) GetKid() int    { return s.Kid }
func (s *TestClaims) SetKid(kid int) { s.Kid = kid }
