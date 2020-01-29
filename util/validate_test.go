package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func checkValidName(t *testing.T, name string, want bool) {
	got := ValidName(name)
	if got != want {
		t.Errorf("checking name %s, wanted %t but got %t",
			name, want, got)
	}
}

func TestValidName(t *testing.T) {
	checkValidName(t, "myname", true)
	checkValidName(t, "my name", true)
	checkValidName(t, "00112", true)
	checkValidName(t, "My_name 0001-0002", true)
	checkValidName(t, "Harry Potter Go", true)
	checkValidName(t, "Deusche Telekom AG", true)
	checkValidName(t, "Telefonica S.A.", true)
	checkValidName(t, "AT&T Inc.", true)
	checkValidName(t, "Niantic, Inc.", true)
	checkValidName(t, "Pokemon Go!", true)
	checkValidName(t, "", false)
	checkValidName(t, " name", false)
	checkValidName(t, "-name", false)
	checkValidName(t, "a;sldfj", false)
	checkValidName(t, "$fadf", false)
}

func checkValidIp(t *testing.T, ip []byte, want bool) {
	got := ValidIp(ip)
	if got != want {
		t.Errorf("checking %x, wanted %t but got %t",
			ip, want, got)
	}
}

func TestValidIp(t *testing.T) {
	checkValidIp(t, []byte{1, 2, 3, 4}, true)
	checkValidIp(t, []byte{1, 2, 3, 4, 5}, false)
	checkValidIp(t, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, 15, 16}, true)
	checkValidIp(t, []byte{1, 2, 3, 4, 5}, false)
	checkValidIp(t, nil, false)
}

func TestValidLDAPName(t *testing.T) {
	checkValidLDAPName(t, "myname", true)
	checkValidLDAPName(t, "my name", true)
	checkValidLDAPName(t, "00112", true)
	checkValidLDAPName(t, "My_name 0001-0002", true)
	checkValidLDAPName(t, "Harry Potter Go", true)
	checkValidLDAPName(t, "Deusche Telekom AG", true)
	checkValidLDAPName(t, "Telefonica S.A.", true)
	checkValidLDAPName(t, "AT&T Inc.", true)
	checkValidLDAPName(t, "Niantic, Inc.", true)
	checkValidLDAPName(t, "Pokemon Go!", true)
	checkValidLDAPName(t, "", false)
	checkValidLDAPName(t, " name", false)
	checkValidLDAPName(t, "name ", false)
	checkValidLDAPName(t, "name\\a", false)
	checkValidLDAPName(t, "name#a", false)
	checkValidLDAPName(t, "name+a", false)
	checkValidLDAPName(t, "name<a", false)
	checkValidLDAPName(t, "name>a", false)
	checkValidLDAPName(t, "name;a", false)
	checkValidLDAPName(t, "name\"a", false)

	name := EscapeLDAPName("foo, Inc.")
	require.Equal(t, "foo, Inc.", UnescapeLDAPName(name))

	user := EscapeLDAPName("jon,user")
	org := EscapeLDAPName("foo, Inc.")
	split := strings.Split(user+","+org, ",")
	require.Equal(t, "jon,user", UnescapeLDAPName(split[0]))
	require.Equal(t, "foo, Inc.", UnescapeLDAPName(split[1]))
}

func checkValidLDAPName(t *testing.T, name string, valid bool) {
	err := ValidLDAPName(name)
	if valid {
		require.Nil(t, err, "name %s should have been valid")
	} else {
		require.NotNil(t, err, "name %s should have been invalid")
	}
}

func TestValidObjName(t *testing.T) {
	var err error

	err = ValidObjName("objname_123.dev")
	require.Nil(t, err, "valid name")
	err = ValidObjName("objname_123$dev")
	require.NotNil(t, err, "invalid name")
	err = ValidObjName("objname_123dev test")
	require.NotNil(t, err, "invalid name")
	err = ValidObjName("objname_123dev,test")
	require.NotNil(t, err, "invalid name")
}

func TestVersion(t *testing.T) {
	var err error

	_, err = VersionParse("2011-10-11")
	require.Nil(t, err, "valid version")

	_, err = VersionParse("2011-30-11")
	require.NotNil(t, err, "invalid version")

	_, err = VersionParse("2011-30-99")
	require.NotNil(t, err, "invalid version")

	_, err = VersionParse("abcd")
	require.NotNil(t, err, "invalid version")

	_, err = VersionParse("20111-11-11")
	require.NotNil(t, err, "invalid version")

	_, err = VersionParse("2011-1-1")
	require.NotNil(t, err, "invalid version")

	err = ValidateImageVersion("2.0.0")
	require.Nil(t, err, "valid image version")

	err = ValidateImageVersion("2.0-0")
	require.Nil(t, err, "valid image version")

	err = ValidateImageVersion("2.0_0")
	require.NotNil(t, err, "invalid image version")

	err = ValidateImageVersion(".2.0.0")
	require.NotNil(t, err, "invalid image version")
}

func TestHeatSanitize(t *testing.T) {
	longstring := make([]rune, 300)
	for i := range longstring {
		longstring[i] = 'a'
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"foo-bar", "foo-bar"},
		{"foo_bar1234567890", "foo_bar1234567890"},
		{"foo.bar-baz_", "foo.bar-baz_"},
		{"foo bar&baz,blah,!no", "foobarbazblahno"},
		{"00foo", "a00foo"},
		{"0jon ber,lin&", "a0jonberlin"},
		{string(longstring), string(longstring[:254])},
	}
	for _, test := range tests {
		str := HeatSanitize(test.name)
		require.Equal(t, test.expected, str)
	}
}
