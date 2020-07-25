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
