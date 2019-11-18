package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

func ValidatePublicKey(pemStr string) (*rsa.PublicKey, error) {

	var err error

	pemBytes := []byte(pemStr)
	pemBlock, rest := pem.Decode(pemBytes)

	if pemBlock != nil && len(rest) == 0 {
		var rsaPubKey interface{}
		switch pemBlock.Type {
		case "PUBLIC KEY":
			rsaPubKey, err = x509.ParsePKIXPublicKey(pemBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("x509.parse pki public key")
			}
		case "RSA PUBLIC KEY":
			rsaPubKey, err = x509.ParsePKCS1PublicKey(pemBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("x509.parse rsa public key")
			}
		default:
			return nil, fmt.Errorf("Unsupported key tpe %q", pemBlock.Type)
		}
		// Assert we got an rsa public key. Returned value is an interface{}
		sshKey, ok := rsaPubKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("Invalid PEM passed")
		}
		return sshKey, nil
	} else if strings.HasPrefix(string(pemBytes), "---- BEGIN SSH2 PUBLIC KEY") {
		// ssh2 public key format (ssh-keygen -m RFC4716)
		// Not supported
		return nil, fmt.Errorf("ssh2 key format not supported")
	} else {
		_, _, _, _, err := ssh.ParseAuthorizedKey(pemBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %v\n", err)
		}
		return nil, nil
	}
}

func ConvertPEMtoOpenSSH(pemStr string) (string, error) {

	sshKey, err := ValidatePublicKey(pemStr)
	if err != nil {
		return "", fmt.Errorf("failed to convert pem key to ssh key: %v", err)
	}
	if err == nil && sshKey == nil {
		// No conversion required
		return pemStr, nil
	}
	// Generate the ssh public key
	pub, err := ssh.NewPublicKey(sshKey)
	if err != nil {
		return "", fmt.Errorf("failed to convert pem key to ssh key: %v", err)
	}

	sshPubKey := base64.StdEncoding.EncodeToString(pub.Marshal())

	return "ssh-rsa " + sshPubKey, nil
}
