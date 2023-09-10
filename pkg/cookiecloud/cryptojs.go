package cookiecloud

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

func Decrypt(passphrase, encryptedData string) ([]byte, error) {
	// base64 decode encrypted data
	encrypted, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("base64 decode err: %w", err)
	}
	salt := encrypted[8:16]
	key_iv, _ := bytesToKey([]byte(passphrase), salt, 48)
	key := key_iv[:32]
	iv := key_iv[32:]
	ciphertext := encrypted[16:]

	return AesDecrypt(ciphertext, key, iv)
}

func Encrypt(passphrase, rawData string) ([]byte, error) {
	salt := make([]byte, 8)
	_, _ = rand.Read(salt)
	key_iv, _ := bytesToKey([]byte(passphrase), salt, 48)
	key := key_iv[:32]
	iv := key_iv[32:]
	encrypted, err := AesEncrypt([]byte(rawData), key, iv)
	if err != nil {
		return nil, err
	}
	encrypted = append([]byte(fmt.Sprintf("Salted__%s", salt)), encrypted...)
	res := base64.StdEncoding.EncodeToString(encrypted)
	return []byte(res), nil
}

func bytesToKey(data []byte, salt []byte, output int) ([]byte, error) {
	if len(salt) != 8 {
		return nil, fmt.Errorf("expected salt of length 8, got %d", len(salt))
	}
	data = append(data, salt...)
	hash := md5.Sum(data)
	key := hash[:]
	finalKey := append([]byte(nil), key...)
	for len(finalKey) < output {
		hash = md5.Sum(append(key, data...))
		key = hash[:]
		finalKey = append(finalKey, key...)
	}
	return finalKey[:output], nil
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data to unpad")
	}
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

func AesEncrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

func AesDecrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
