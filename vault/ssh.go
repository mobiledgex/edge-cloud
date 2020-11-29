package vault

import "fmt"

func SignSSHKey(config *Config, publicKey string) (string, error) {
	data := map[string]interface{}{
		"public_key": publicKey,
	}
	client, err := config.Login()
	if err != nil {
		return "", err
	}
	ssh := client.SSH()
	secret, err := ssh.SignKey("machine", data)
	if err != nil {
		return "", err
	}
	signedKey, ok := secret.Data["signed_key"]
	if !ok {
		return "", fmt.Errorf("failed to get signed key from vault: %v", secret)
	}
	signedKeyStr, ok := signedKey.(string)
	if !ok {
		return "", fmt.Errorf("invalid signed key from vault: %v", signedKey)
	}
	return signedKeyStr, nil
}

type MEXKey struct {
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
}
