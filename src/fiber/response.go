package fiber

const UNAUTHORIZED_ERROR_CODE string = "00400001"
const UNAUTHORIZED_ERROR_DESC string = "Apikey is missing or invalid"
const INVALID_SIGNATURE_ERROR_CODE string = "00400002"
const INVALID_SIGNATURE_ERROR_DESC string = "Invalid signature"

type APIError struct {
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}
