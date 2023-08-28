package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LoggingMessage struct {
	TransactionTime   string `json:"tranTime" example:"2022-06-14T17:09:05.556+07:00"`
	RequestId         string `json:"requestID" example:"8e5b2ae9-2db7-4fd6-9f51-dfe303e59719"`
	UserAgent         string `json:"userAgent" example:"Mozilla/5.0 (Linux; Android 7.0; SM-G930V Build/NRD90M) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.125 Mobile Safari/537.36"`
	RemoteAddress     string `json:"remoteAddress" example:"223.205.245.198"`
	PartnerId         string `json:"partnerID" example:"ABC"`
	Service           string `json:"service" example:"authen-service"`
	RawUrl            string `json:"rawURL" example:"https://abc.com/path/to/api"`
	HttpMethod        string `json:"httpMethod" example:"POST"`
	HttpStatus        string `json:"httpStatus" example:"503"`
	ErrorType         string `json:"ErrorType" example:"SystemError"`
	ExecutionTimeMsec uint64 `json:"executionTime" example:"55"`
	ResponseBody      string `json:"responseBody" example:"{\"code\":\"8001\", \"title\":\"service is not available at the moment\"}"`
	StackTrace        string `json:"stackTrace" example:"Exception in thread \"main\" java.lang.NullPointerException"`
}

type Logger struct {
	StartTime time.Time
	Msg       LoggingMessage
}

func (logger *Logger) Intialize(opts ...Option) {
	def := options{
		RequestId:     "",
		UserAgent:     "",
		RemoteAddress: "",
		PartnerId:     "",
		Service:       "",
		RawUrl:        "",
		HttpMethod:    LOGGING_HTTPMETHOD_GET,
		HttpStatus:    LOGGING_HTTPSTATUS_OK,
		ErrorType:     LOGGING_ERRORTYPE_NONE,
	}

	for _, opt := range opts {
		opt.set(&def)
	}

	now := time.Now()
	logger.StartTime = now
	logger.Msg.TransactionTime = now.Format("2006-01-02T15:04:05.000") + now.Format("-07:00")
	logger.Msg.RequestId = def.RequestId
	logger.Msg.UserAgent = def.UserAgent
	logger.Msg.RemoteAddress = def.RemoteAddress
	logger.Msg.PartnerId = def.PartnerId
	logger.Msg.Service = def.Service
	logger.Msg.RawUrl = def.RawUrl
	logger.Msg.HttpMethod = def.HttpMethod
	logger.Msg.HttpStatus = def.HttpStatus
	logger.Msg.ErrorType = def.ErrorType
	logger.Msg.ExecutionTimeMsec = 0
	logger.Msg.ResponseBody = ""
	logger.Msg.StackTrace = ""
}

func (logger *Logger) Print(ctx *fiber.Ctx) {
	if ctx != nil {
		other := ctx.Locals("logger").(*Logger)
		logger.Msg = other.Msg
		logger.Msg.HttpStatus = fmt.Sprintf("%d", ctx.Response().StatusCode())
		logger.Msg.ResponseBody = string(ctx.Response().Body())
	}

	now := time.Now()
	logger.Msg.ExecutionTimeMsec = uint64(now.Sub(logger.StartTime).Milliseconds())

	raw, _ := json.Marshal(logger.Msg)
	log.SetFlags(0)
	log.Print(string(raw))
}
