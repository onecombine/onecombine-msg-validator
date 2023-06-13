package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewStruct(key string, age int32) Validator {
	h := NewOneCombineHmac(key, age)
	return h.(Validator)
}

func TestOneCombineHmacBasic(t *testing.T) {
	hmc := NewStruct("hello", 100000)
	assert.NotEqual(t, "", hmc.Sign("{\"a\":\"yes\",\"b\":5}"))

	assert.Equal(t, "", hmc.Sign("{\"partner_id\":\"500001\", \"payee\":\"payeeliquid\", \"transaction_datetime\":\"2018-08-07T10:00:00+08:00\",\"crn\":\"12345\", \"currency_code\":\"SGD\", \"amount\":\"0.01\", \"channel\":\"00\", \"channel_info\":\"134631414264156089\"}"))
}
