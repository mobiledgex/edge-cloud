package util

import (
	"testing"
)

var (
	ssh_public_key string = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrHlOJOJUqvd4nEOXQbdL8ODKzWaUxKVY94pF7J3diTxgZ1NTvS6omqOjRS3loiU7TOlQQU4cKnRRnmJW8QQQZSOMIGNrMMInGaEYsdm6+tr1k4DDfoOrkGMj3X/I2zXZ3U+pDPearVFbczCByPU0dqs16TWikxDoCCxJRGeeUl7duzD9a65bI8Jl+zpfQV+I7OPa81P5/fw15lTzT4+F9MhhOUVJ4PFfD+d6/BLnlUfZ94nZlvSYnT+GoZ8xTAstM7+6pvvvHtaHoV4YqRf5CelbWAQ162XNa9/pW5v/RKDrt203/JEk3e70tzx9KAfSw2vuO1QepkCZAdM9rQoCd"

	ssh2_public_key string = `---- BEGIN SSH2 PUBLIC KEY ----
AAAAB3NzaC1yc2EAAAADAQABAAABAQCrHlOJOJUqvd4nEOXQbdL8ODKzWaUxKVY94pF7J3
diTxgZ1NTvS6omqOjRS3loiU7TOlQQU4cKnRRnmJW8QQQZSOMIGNrMMInGaEYsdm6+tr1k
4DDfoOrkGMj3X/I2zXZ3U+pDPearVFbczCByPU0dqs16TWikxDoCCxJRGeeUl7duzD9a65
bI8Jl+zpfQV+I7OPa81P5/fw15lTzT4+F9MhhOUVJ4PFfD+d6/BLnlUfZ94nZlvSYnT+Go
Z8xTAstM7+6pvvvHtaHoV4YqRf5CelbWAQ162XNa9/pW5v/RKDrt203/JEk3e70tzx9KAf
Sw2vuO1QepkCZAdM9rQoCd
---- END SSH2 PUBLIC KEY ----`

	pem_public_key string = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqx5TiTiVKr3eJxDl0G3S
/Dgys1mlMSlWPeKReyd3Yk8YGdTU70uqJqjo0Ut5aIlO0zpUEFOHCp0UZ5iVvEEE
GUjjCBjazDCJxmhGLHZuvra9ZOAw36Dq5BjI91/yNs12d1PqQz3mq1RW3Mwgcj1N
HarNek1opMQ6AgsSURnnlJe3bsw/WuuWyPCZfs6X0FfiOzj2vNT+f38NeZU80+Ph
fTIYTlFSeDxXw/nevwS55VH2feJ2Zb0mJ0/hqGfMUwLLTO/uqb77x7Wh6FeGKkX+
QnpW1gENetlzWvf6Vub/0Sg67dtN/yRJN3u9Lc8fSgH0sNr7jtUHqZAmQHTPa0KA
nQIDAQAB
-----END PUBLIC KEY-----`

	pem_rsa_public_key string = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAqx5TiTiVKr3eJxDl0G3S/Dgys1mlMSlWPeKReyd3Yk8YGdTU70uq
Jqjo0Ut5aIlO0zpUEFOHCp0UZ5iVvEEEGUjjCBjazDCJxmhGLHZuvra9ZOAw36Dq
5BjI91/yNs12d1PqQz3mq1RW3Mwgcj1NHarNek1opMQ6AgsSURnnlJe3bsw/WuuW
yPCZfs6X0FfiOzj2vNT+f38NeZU80+PhfTIYTlFSeDxXw/nevwS55VH2feJ2Zb0m
J0/hqGfMUwLLTO/uqb77x7Wh6FeGKkX+QnpW1gENetlzWvf6Vub/0Sg67dtN/yRJ
N3u9Lc8fSgH0sNr7jtUHqZAmQHTPa0KAnQIDAQAB
-----END RSA PUBLIC KEY-----`
)

func TestPublicKeyAPI(t *testing.T) {
	_, err := ValidatePublicKey(ssh_public_key)
	if err != nil {
		t.Errorf("validation error %v", err)
	}

	_, err = ValidatePublicKey(ssh2_public_key)
	if err == nil {
		t.Errorf("validation error ssh2 should not be supported")
	}

	_, err = ValidatePublicKey(pem_public_key)
	if err != nil {
		t.Errorf("validation error %v", err)
	}

	_, err = ValidatePublicKey(pem_rsa_public_key)
	if err != nil {
		t.Errorf("validation error %v", err)
	}

	rKey, err := ConvertPEMtoOpenSSH(ssh_public_key)
	if err != nil || rKey != ssh_public_key {
		t.Errorf("conversion error %v", err)
	}

	rKey, err = ConvertPEMtoOpenSSH(pem_public_key)
	if err != nil || rKey != ssh_public_key {
		t.Errorf("conversion error %v", err)
	}

	rKey, err = ConvertPEMtoOpenSSH(pem_rsa_public_key)
	if err != nil || rKey != ssh_public_key {
		t.Errorf("conversion error %v", err)
	}
}
