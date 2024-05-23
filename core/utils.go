package core

import (
	"encoding/base64"
	"encoding/hex"
)

func DecodeBase64(base64String string) ([]byte, error) {
	p, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func EncodeToHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func Base64ToHex(base64String string) (string, error) {
	bytes, err := DecodeBase64(base64String)
	if err != nil {
		return "", err
	}
	return EncodeToHex(bytes), nil
}

func Ptr[T any](value T) *T {
	return &value
}
