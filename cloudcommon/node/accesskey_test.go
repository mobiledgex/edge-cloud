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
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

func TestKeyPair(t *testing.T) {
	pair, err := GenerateAccessKey()
	require.Nil(t, err)

	fmt.Printf("private key:\n%s\n", pair.PrivatePEM)
	fmt.Printf("public key:\n%s\n", pair.PublicPEM)

	privKey, err := LoadPrivPEM([]byte(pair.PrivatePEM))
	require.Nil(t, err)
	pubKey, err := LoadPubPEM([]byte(pair.PublicPEM))
	require.Nil(t, err)

	msg := "here's some message"
	sig := ed25519.Sign(privKey, []byte(msg))
	sigb64 := base64.StdEncoding.EncodeToString(sig)

	fmt.Printf("signature: %s\n", sigb64)

	sig, err = base64.StdEncoding.DecodeString(sigb64)
	require.Nil(t, err)

	ok := ed25519.Verify(pubKey, []byte(msg), sig)
	require.True(t, ok)
}
