package cookiecloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {
	data := "this is data"
	key := "this is key"

	encryptedData, err := Encrypt(key, data)
	assert.Nil(t, err)
	decryptedData, err := Decrypt(key, string(encryptedData))
	assert.Nil(t, err)

	assert.Equal(t, data, string(decryptedData))
}
