package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Run("encrypt -> decrypt", func(t *testing.T) {
		encrypted, err := EncryptToken("this is my token", "password")
		assert.NoError(t, err)

		decrypted, err := DecryptToken(encrypted, "password")
		assert.NoError(t, err)

		assert.Equal(t, "this is my token", decrypted)
	})
}
