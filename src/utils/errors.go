package utils

import "encoding/json"

const (
	CODE_INTERNAL_ERROR    = "00500001"
	CODE_APIKEY_MISSING    = "00400001"
	CODE_INVALID_SIGNATURE = "00400002"
	CODE_BAD_REQUEST       = "00400006"
	CODE_ORDER_NOT_FOUND   = "00404001"
	CODE_ORDER_REF_EXIST   = "00400009"

	// Reversal
	CODE_REFUND_NOT_ALLOW         = "20402002"
	CODE_CANCEL_NOT_ALLOW         = "20402003"
	CODE_CURRENCY_NOT_MATCH_ORGTX = "20402005"
	CODE_AMOUNT_EXCEEDED_ORGTX    = "20402006"
)

const (
	MSG_INTERNAL_ERROR    = "Internal system error"
	MSG_APIKEY_MISSING    = "Apikey is missing or invalid"
	MSG_INVALID_SIGNATURE = "Invalid signature"
	MSG_BAD_REQUEST       = "A field contains invalid value"
	MSG_ORDER_NOT_FOUND   = "Order cannot be found"
	MSG_ORDER_REF_EXIST   = "order_ref already exists"

	// Reversal
	MSG_REFUND_NOT_ALLOW         = "Refund transaction operation is not allow"
	MSG_CANCEL_NOT_ALLOW         = "Cancel transaction operation is not allowed"
	MSG_CURRENCY_NOT_MATCH_ORGTX = "Refund/Cancel currency does not match transaction currency"
	MSG_AMOUNT_EXCEEDED_ORGTX    = "Refund/Cancel amount exceeds remaining amount that can be refunded/cancelled"
)

var (
	errorDict = map[string]string{
		CODE_INTERNAL_ERROR:           MSG_INTERNAL_ERROR,
		CODE_APIKEY_MISSING:           MSG_APIKEY_MISSING,
		CODE_INVALID_SIGNATURE:        MSG_INVALID_SIGNATURE,
		CODE_BAD_REQUEST:              MSG_BAD_REQUEST,
		CODE_ORDER_NOT_FOUND:          MSG_ORDER_NOT_FOUND,
		CODE_REFUND_NOT_ALLOW:         MSG_REFUND_NOT_ALLOW,
		CODE_CANCEL_NOT_ALLOW:         MSG_CANCEL_NOT_ALLOW,
		CODE_CURRENCY_NOT_MATCH_ORGTX: MSG_CURRENCY_NOT_MATCH_ORGTX,
		CODE_AMOUNT_EXCEEDED_ORGTX:    MSG_AMOUNT_EXCEEDED_ORGTX,
	}
)

type ErrorResponse struct {
	Code        string `json:"error_code"`
	Description string `json:"error_description"`
}

// General Error Response (frequently used)
func BadRequestError() *ErrorResponse {
	return &ErrorResponse{
		Code:        CODE_BAD_REQUEST,
		Description: MSG_BAD_REQUEST,
	}
}

func InternalSystemError() *ErrorResponse {
	return &ErrorResponse{
		Code:        CODE_INTERNAL_ERROR,
		Description: MSG_INTERNAL_ERROR,
	}
}

func APIKeyError() *ErrorResponse {
	return &ErrorResponse{
		Code:        CODE_APIKEY_MISSING,
		Description: MSG_APIKEY_MISSING,
	}
}

func InvalidSignature() *ErrorResponse {
	return &ErrorResponse{
		Code:        CODE_INVALID_SIGNATURE,
		Description: MSG_INVALID_SIGNATURE,
	}
}

// General purpose error
func CreateErrorResponse(code string) *ErrorResponse {

	desc, ok := errorDict[code]
	if ok {
		return &ErrorResponse{
			Code:        code,
			Description: desc,
		}
	}

	return nil
}

func ErrorResponseToMap(e *ErrorResponse) map[string]string {
	out := make(map[string]string)
	j, _ := json.Marshal(e)
	json.Unmarshal(j, &out)
	return out
}
