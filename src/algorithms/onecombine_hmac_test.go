package algorithms

import (
	"fmt"
	"testing"

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
	tstamp := "1687227085"
	signature := "t=" + tstamp + "," + "sKcS/n+FhIXnHxdJZZAsn+mRCF6t046rxh47SaxvzhY="
	assert.Equal(t, true, hmc.Verify([]byte(jmsg), signature))
}

func TestOneCombineHmacVerify01(t *testing.T) {
	// Wrong secret
	hmc := NewStruct("hellooooo", 100000)
	jmsg := "{\"partner_id\":\"500001\", \"payee\":\"payeeliquid\", \"transaction_datetime\":\"2018-08-07T10:00:00+08:00\",\"crn\":\"12345\", \"currency_code\":\"SGD\", \"amount\":\"0.01\", \"channel\":\"00\", \"channel_info\":\"134631414264156089\"}"
	tstamp := "1687227085"
	signature := "t=" + tstamp + "," + "sKcS/n+FhIXnHxdJZZAsn+mRCF6t046rxh47SaxvzhY="
	assert.Equal(t, false, hmc.Verify([]byte(jmsg), signature))
}
