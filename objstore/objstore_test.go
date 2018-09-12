package objstore

import "testing"
import "github.com/stretchr/testify/assert"

func TestDbKeyPrefixParse(t *testing.T) {
	reg, typ, key, err := DbKeyPrefixParse("1/app/my app key")
	assert.Nil(t, err)
	assert.Equal(t, "1", reg)
	assert.Equal(t, "app", typ)
	assert.Equal(t, "my app key", key)

	reg, typ, key, err = DbKeyPrefixParse("1/app")
	assert.NotNil(t, err)
}
