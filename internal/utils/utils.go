package utils

import (
	"golang.org/x/crypto/ssh"
)

func GenerateFingerprint(key []byte) (string, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return "", err
	}

	fingerprintSHA256 := ssh.FingerprintSHA256(pubKey)

	return fingerprintSHA256, nil
}
