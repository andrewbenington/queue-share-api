package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/andrewbenington/queue-share-api/config"
)

func hashTo64(value string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(value))
	return hasher.Sum(nil)
}

func AESGCMEncrypt(token string) ([]byte, error) {
	keyBytes := hashTo64(config.GetEncryptionKey())
	tokenBytes := []byte(token)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, tokenBytes, nil)
	return ciphertext, nil
}

func AESGCMDecrypt(encrypted []byte) (string, error) {
	key := hashTo64(config.GetEncryptionKey())

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
