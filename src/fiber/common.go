package fiber

import (
	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
)

const MESSAGE_EXPIRATION_MSEC string = "MESSAGE_EXPIRATION_MSEC"

type XnapUtility struct {
	ApiKey    string
	Validator *algorithms.Validator
}

type AcquirerUtility struct {
	validator *algorithms.Validator
	id        string
}
