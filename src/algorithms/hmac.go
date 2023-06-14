package algorithms

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type HmacSha256 struct {
	Key []byte
}

func NewHmacSha256(key string) *HmacSha256 {
	var instance HmacSha256
	instance.Key = []byte(key)
	return &instance
}

func (hmc HmacSha256) Sign(data []byte, options ...string) string {
	hash := hmac.New(sha256.New, hmc.Key)
	hash.Write(data)
	sha := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return sha
}

func (hmc HmacSha256) Verify(data []byte, signature string) bool {
	sig := hmc.Sign(data)

	return sig == signature
}
