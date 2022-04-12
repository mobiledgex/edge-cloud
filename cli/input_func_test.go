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

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveAlias(t *testing.T) {
	aliases := map[string]string{
		"inner:#.name": "foo.inner:#.name",
		"a:#.b:#.name": "arr:#.innerarr:#.name",
	}
	reals := make(map[string]string)
	for k, v := range aliases {
		reals[v] = k
	}

	test := func(alias, real string) {
		require.Equal(t, real, resolveAlias(alias, aliases))
		require.Equal(t, alias, resolveAlias(real, reals))
	}

	test("inner:0.name", "foo.inner:0.name")
	test("a:2.b:190.name", "arr:2.innerarr:190.name")
}
