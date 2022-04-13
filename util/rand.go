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

import "math/rand"

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandAscii(n int) []byte {
	output := make([]byte, n)
	random := make([]byte, n)
	_, err := rand.Read(random)
	l := len(letterBytes)
	if err == nil {
		for ii := 0; ii < n; ii++ {
			randPos := uint8(random[ii]) % uint8(l)
			output[ii] = letterBytes[randPos]
		}
	} else {
		// slower
		for ii := 0; ii < n; ii++ {
			randPos := uint8(rand.Int63()) % uint8(l)
			output[ii] = letterBytes[randPos]
		}
	}
	return output
}
