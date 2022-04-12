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

package objstore

import "testing"
import "github.com/stretchr/testify/assert"

func TestDbKeyPrefixParse(t *testing.T) {
	reg, typ, key, err := DbKeyPrefixParse("1/app/my app key")
	assert.Nil(t, err)
	assert.Equal(t, "1", reg)
	assert.Equal(t, "app", typ)
	assert.Equal(t, "my app key", key)

	reg, typ, key, err = DbKeyPrefixParse("1/version")
	assert.Nil(t, err)
}
