package util

import "testing"

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
