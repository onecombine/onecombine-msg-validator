package src

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type HmacSha256 struct {
	Key []byte
}

func NewHmacSha256(key string) *HmacSha256 {
	var instance HmacSha256
	instance.Key = []byte(key)
	return &instance
}

func (hmc HmacSha256) Sign(data string) string {
	hash := hmac.New(sha256.New, hmc.Key)
	hash.Write([]byte(data))
	sha := hex.EncodeToString(hash.Sum(nil))
	return sha
}

func (hmc HmacSha256) Verify(data, signature string) bool {
	sig := hmc.Sign(data)

	return sig == signature
}
