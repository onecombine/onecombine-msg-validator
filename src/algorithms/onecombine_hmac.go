package algorithms

import (
	"math"
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

func (oc OneCombineHmac) Reformat(data []byte, timestamp string) string {
	var tstamp string
	if tstamp = timestamp; len(timestamp) == 0 {
		now := time.Now().Unix()
		tstamp = strconv.FormatInt(now, 10)
	}
	filtered := ""
	for _, ch := range data {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '{' || ch == '}' || ch == ':' || ch == ',' || ch == '.' {
			filtered += string(ch)
		}
	}
	filtered += ":"
	filtered += tstamp

	return strings.ToUpper(filtered)
}

func (oc OneCombineHmac) Sign(data string, options ...string) string {
	tstamp := ""
	if len(options) > 0 {
		tstamp = options[0]
	}
	filtered := oc.Reformat([]byte(data), tstamp)
	return oc.Hmac.Sign([]byte(filtered))
}

func (oc OneCombineHmac) Verify(data []byte, signature string) bool {
	prefix := signature[0:2]
	if prefix != "t=" || !strings.Contains(signature, ",") {
		return false
	}
	signature = signature[2:]

	parts := strings.Split(signature, ",")
	tstamp := parts[0]
	signature = parts[1]

	tstampValue, err := strconv.ParseInt(tstamp, 10, 64)
	now := time.Now().Unix()
	if err != nil || math.Abs(float64(tstampValue)-float64(now)) > float64(oc.MaxAge) {
		return false
	}

	filtered := oc.Reformat(data, tstamp)
	return oc.Hmac.Verify([]byte(filtered), signature)
}
