package acceptance

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

func generateKeyMaterial(publicKeyB64 string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "", fmt.Errorf("error decoding public key: %s", err)
	}

	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("error parsing public key: %s", err)
	}

	rsaPublicKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not RSA")
	}

	keyMaterial := make([]byte, 32)
	if _, err := rand.Read(keyMaterial); err != nil {
		return "", fmt.Errorf("error generating random key material: %s", err)
	}

	encryptedKeyMaterial, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, keyMaterial)
	if err != nil {
		return "", fmt.Errorf("error encrypting key material: %s", err)
	}

	return base64.StdEncoding.EncodeToString(encryptedKeyMaterial), nil
}
