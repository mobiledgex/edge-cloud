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
	checkValidName(t, "Hunna Stoll Go", true)
	checkValidName(t, "Deusche telecom AG", true)
	checkValidName(t, "Sonoral S.A.", true)
	checkValidName(t, "UFGT Inc.", true)
	checkValidName(t, "Atlantic, Inc.", true)
	checkValidName(t, "Pillimo Go!", true)
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
	checkValidLDAPName(t, "Hunna Stoll Go", true)
	checkValidLDAPName(t, "Deusche telecom AG", true)
	checkValidLDAPName(t, "Sonoral S.A.", true)
	checkValidLDAPName(t, "UFGT Inc.", true)
	checkValidLDAPName(t, "Atlantic, Inc.", true)
	checkValidLDAPName(t, "Pillimo Go!", true)
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

	_, err = ContainerVersionParse("2011-10-11")
	require.Nil(t, err, "valid version")

	_, err = ContainerVersionParse("2011-30-11")
	require.NotNil(t, err, "invalid version")

	_, err = ContainerVersionParse("2011-30-99")
	require.NotNil(t, err, "invalid version")

	_, err = ContainerVersionParse("abcd")
	require.NotNil(t, err, "invalid version")

	_, err = ContainerVersionParse("20111-11-11")
	require.NotNil(t, err, "invalid version")

	_, err = ContainerVersionParse("2011-1-1")
	require.NotNil(t, err, "invalid version")

	err = ValidateImageVersion("2.0.0")
	require.Nil(t, err, "valid image version")

	err = ValidateImageVersion("2.0-0")
	require.Nil(t, err, "valid image version")

	err = ValidateImageVersion("2.0_0")
	require.Nil(t, err, "valid image version")

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
		{"0jon ber,lin&", "a0jonbeacon"},
		{string(longstring), string(longstring[:254])},
	}
	for _, test := range tests {
		str := HeatSanitize(test.name)
		require.Equal(t, test.expected, str)
	}
}

func TestImagePath(t *testing.T) {
	validPaths := []string{
		"https://artifactory-qa.mobiledgex.net/artifactory/repo-MobiledgeX/server_ping_threaded_centos7.qcow2#md5:5ce8dbcdd8b7c2054779d742f4bf602d",
	}
	for _, imgPath := range validPaths {
		err := ValidateImagePath(imgPath)
		require.Nil(t, err, "valid image path")
	}
	invalidPaths := []string{
		"https://artifactory-qa.mobiledgex.net/artifactory/repo-MobiledgeX#md5:5ce8dbcdd8b7c2054779d742f4bf602d",
		"https://artifactory-qa.mobiledgex.net/artifactory/repo-MobiledgeX/abc.qcow2",
	}
	for _, imgPath := range invalidPaths {
		err := ValidateImagePath(imgPath)
		require.NotNil(t, err, "invalid image path")
	}
}

func TestK8SContainerName(t *testing.T) {
	validNames := []string{
		"testapp-12334",
		"testpod-123/testcontainer-1234",
	}
	inValidNames := []string{
		"testapp-12334; rm -rf .",
		"testpod-123/testcontainer-1234 && rm -rf .",
	}
	for _, name := range validNames {
		err := ValidK8SContainerName(name)
		require.Nil(t, err, "valid k8s container name")
	}
	for _, name := range inValidNames {
		err := ValidK8SContainerName(name)
		require.NotNil(t, err, "invalid k8s container name")
	}
}
