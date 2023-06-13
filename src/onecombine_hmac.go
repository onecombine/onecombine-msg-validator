package src

import (
	"strconv"
	"strings"
	"time"
)

type OneCombineHmac struct {
	Hmac   *HmacSha256
	MaxAge int32
}

func NewOneCombineHmac(key string, maxAge int32) interface{} {
	var instance OneCombineHmac
	instance.Hmac = NewHmacSha256(key)
	instance.MaxAge = maxAge
	return &instance
}

func (oc OneCombineHmac) Reformat(data string, timestamp string) string {
	var tstamp string
	if tstamp = timestamp; len(timestamp) == 0 {
		now := time.Now().Unix()
		tstamp = strconv.FormatInt(now, 10)
	}
	filtered := ""
	prev := ""
	for _, ch := range data {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '{' || ch == '}' || ch == ':' || ch == ',' || ch == '.' {
			filtered += prev
			prev = string(ch)
		}
	}
	filtered += ":"
	filtered += tstamp
	filtered += "}"

	return strings.ToUpper(filtered)
}

func (oc OneCombineHmac) Sign(data string) string {
	filtered := oc.Reformat(data, "")
	return oc.Hmac.Sign(filtered)
}

func (oc OneCombineHmac) Verify(data, signature string) bool {
	prefix := signature[0:2]
	if prefix != "t=" || !strings.Contains(signature, ",") {
		return false
	}
	signature = signature[2:]

	parts := strings.Split(signature, ",")
	tstamp := parts[0]
	signature = parts[1]

	filtered := oc.Reformat(data, tstamp)
	return oc.Hmac.Verify(filtered, signature)
}
