package utils

type options struct {
	RequestId     string
	DeviceId      string
	UserAgent     string
	RemoteAddress string
	UserId        string
	ChannelId     string
	Service       string
	RawUrl        string
	HttpMethod    string
	HttpStatus    string
	ErrorType     string
}

type Option interface {
	set(*options)
}

type LoggingRequestId string
type LoggingDeviceId string
type LoggingUserAgent string
type LoggingRemoteAddress string
type LoggingUserId string
type LoggingChannelId string
type LoggingService string
type LoggingRawUrl string
type LoggingHttpMethod string
type LoggingHttpStatus string
type LoggingErrorType string

func (v LoggingRequestId) set(opts *options) {
	opts.RequestId = string(v)
}
func (v LoggingDeviceId) set(opts *options) {
	opts.DeviceId = string(v)
}
func (v LoggingUserAgent) set(opts *options) {
	opts.UserAgent = string(v)
}
func (v LoggingRemoteAddress) set(opts *options) {
	opts.RemoteAddress = string(v)
}
func (v LoggingUserId) set(opts *options) {
	opts.UserId = string(v)
}
func (v LoggingChannelId) set(opts *options) {
	opts.ChannelId = string(v)
}
func (v LoggingService) set(opts *options) {
	opts.Service = string(v)
}
func (v LoggingRawUrl) set(opts *options) {
	opts.RawUrl = string(v)
}
func (v LoggingHttpMethod) set(opts *options) {
	opts.HttpMethod = string(v)
}
func (v LoggingHttpStatus) set(opts *options) {
	opts.HttpStatus = string(v)
}
func (v LoggingErrorType) set(opts *options) {
	opts.ErrorType = string(v)
}

func WithRequestId(v string) Option {
	return LoggingRequestId(v)
}
func WithDeviceId(v string) Option {
	return LoggingDeviceId(v)
}
func WithUserAgent(v string) Option {
	return LoggingUserAgent(v)
}
func WithRemoteAddress(v string) Option {
	return LoggingRemoteAddress(v)
}
func WithUserId(v string) Option {
	return LoggingUserId(v)
}
func WithChannelId(v string) Option {
	return LoggingChannelId(v)
}
func WithService(v string) Option {
	return LoggingService(v)
}
func WithRawUrl(v string) Option {
	return LoggingRawUrl(v)
}

const LOGGING_HTTPMETHOD_POST string = "POST"
const LOGGING_HTTPMETHOD_GET string = "GET"
const LOGGING_HTTPMETHOD_PUT string = "PUT"
const LOGGING_HTTPMETHOD_DELETE string = "DELETE"

func WithHttpMethod(v string) Option {
	return LoggingHttpMethod(v)
}

const LOGGING_HTTPSTATUS_OK string = "200"
const LOGGING_HTTPSTATUS_CREATED string = "201"
const LOGGING_HTTPSTATUS_ACCEPTED string = "202"
const LOGGING_HTTPSTATUS_NOCONTENT string = "204"
const LOGGING_HTTPSTATUS_BADREQUEST string = "400"
const LOGGING_HTTPSTATUS_UNAUTHORIZED string = "401"
const LOGGING_HTTPSTATUS_FORBIDDEN string = "403"
const LOGGING_HTTPSTATUS_NOTFOUND string = "404"
const LOGGING_HTTPSTATUS_METHODNOTALLOWED string = "405"
const LOGGING_HTTPSTATUS_NOTACCEPTABLE string = "406"
const LOGGING_HTTPSTATUS_CONFLICT string = "409"
const LOGGING_HTTPSTATUS_INTERNALSERVERERROR string = "500"
const LOGGING_HTTPSTATUS_NOTIMPLEMENTED string = "501"
const LOGGING_HTTPSTATUS_SERVICEUNAVALABLE string = "503"

func WithHttpStatus(v string) Option {
	return LoggingHttpMethod(v)
}

const LOGGING_ERRORTYPE_SYSTEMERROR string = "SystemError"
const LOGGING_ERRORTYPE_BUSINESSERROR string = "BusinessError"

func WithErrorType(v string) Option {
	return LoggingErrorType(v)
}
