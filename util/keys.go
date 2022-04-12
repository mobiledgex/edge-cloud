// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
			return nil, fmt.Errorf("invalid PEM passed")
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
