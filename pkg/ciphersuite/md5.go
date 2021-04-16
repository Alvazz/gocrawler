package ciphersuite

import (
	"crypto/md5"
	"encoding/hex"
)

// GetMD5Hash obtiene el hash MD5 de string text
func GetMD5Hash(text string) (string, error) {
	encryptor := md5.New()
	if _, err := encryptor.Write([]byte(text)); err != nil {
		return "", err
	}
	return hex.EncodeToString(encryptor.Sum(nil)), nil
}
