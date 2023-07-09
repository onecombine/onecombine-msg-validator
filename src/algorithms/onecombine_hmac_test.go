package algorithms

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func NewStruct(key string, age int32) Validator {
	h := NewOneCombineHmac(key, age)
	return h.(Validator)
}

func TestOneCombineHmacBasic(t *testing.T) {
	hmc := NewStruct("hello", 100000)
	jmsg := "{\"partner_id\":\"500001\", \"payee\":\"payeeliquid\", \"transaction_datetime\":\"2018-08-07T10:00:00+08:00\",\"crn\":\"12345\", \"currency_code\":\"SGD\", \"amount\":\"0.01\", \"channel\":\"00\", \"channel_info\":\"134631414264156089\"}"
	tstamp := "1687227085"
	assert.Equal(t, fmt.Sprintf("t=%s,", tstamp)+"sKcS/n+FhIXnHxdJZZAsn+mRCF6t046rxh47SaxvzhY=", hmc.Sign(jmsg, tstamp))
}

func TestOneCombineHmacVerify00(t *testing.T) {
	// All is well
	hmc := NewStruct("hello", 100000)
	jmsg := "{\"partner_id\":\"500001\", \"payee\":\"payeeliquid\", \"transaction_datetime\":\"2018-08-07T10:00:00+08:00\",\"crn\":\"12345\", \"currency_code\":\"SGD\", \"amount\":\"0.01\", \"channel\":\"00\", \"channel_info\":\"134631414264156089\"}"
	now := math.Floor((float64(time.Now().UnixMilli()) / 1000))
	tstamp := fmt.Sprintf("%d", int64(now))
	sig := hmc.Sign(jmsg, tstamp)
	assert.Equal(t, true, hmc.Verify([]byte(jmsg), sig))
}

func TestOneCombineHmacVerify01(t *testing.T) {
	// Wrong secret
	hmc := NewStruct("hello", 100000)
	jmsg := "{\"partner_id\":\"500001\", \"payee\":\"payeeliquid\", \"transaction_datetime\":\"2018-08-07T10:00:00+08:00\",\"crn\":\"12345\", \"currency_code\":\"SGD\", \"amount\":\"0.01\", \"channel\":\"00\", \"channel_info\":\"134631414264156089\"}"
	now := math.Floor((float64(time.Now().UnixMilli()) / 1000))
	tstamp := fmt.Sprintf("%d", int64(now))
	sig := hmc.Sign(jmsg, tstamp)

	hmc2 := NewStruct("hellooooo", 100000)
	assert.Equal(t, false, hmc2.Verify([]byte(jmsg), sig))
}
