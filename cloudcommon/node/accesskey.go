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

package node

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ed25519"
)

type KeyPair struct {
	PrivatePEM string
	PublicPEM  string
}

// Generate ed25519 key pair
func GenerateAccessKey() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	pair := KeyPair{}

	// convert private key to pem
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	privBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	pair.PrivatePEM = string(pem.EncodeToMemory(&privBlock))

	// convert public key to pem
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	pubBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	pair.PublicPEM = string(pem.EncodeToMemory(&pubBlock))
	return &pair, nil
}

func LoadPrivPEM(key []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("No key found in pem string")
	}
	dat, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := dat.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("Key not ed25519 format")
	}
	return k, nil
}

func LoadPubPEM(key []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("No key found in pem string")
	}
	dat, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := dat.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Key not ed25519 format")
	}
	return k, nil
}
