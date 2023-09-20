package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Run("encrypt -> decrypt", func(t *testing.T) {
		t.Setenv("ENCRYPTION_KEY", "password")
		encrypted, err := AESGCMEncrypt("this is my token")
		assert.NoError(t, err)

		decrypted, err := AESGCMDecrypt(encrypted, "password")
		assert.NoError(t, err)

		assert.Equal(t, "this is my token", decrypted)
	})
}
