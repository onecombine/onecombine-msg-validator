package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestInitialize(t *testing.T) {
	logger := Logger{}
	logger.Intialize()

	expected := Logger{}
	expected.StartTime = logger.StartTime
	expected.Msg.TransactionTime = logger.Msg.TransactionTime
	expected.Msg.RequestId = ""
	expected.Msg.UserAgent = ""
	expected.Msg.RemoteAddress = ""
	expected.Msg.PartnerId = ""
	expected.Msg.Service = ""
	expected.Msg.RawUrl = ""
	expected.Msg.HttpMethod = LOGGING_HTTPMETHOD_GET
	expected.Msg.HttpStatus = LOGGING_HTTPSTATUS_OK
	expected.Msg.ErrorType = LOGGING_ERRORTYPE_NONE
	expected.Msg.ExecutionTimeMsec = 0
	expected.Msg.ResponseBody = ""
	expected.Msg.StackTrace = ""
	assert.Equal(t, expected, logger, "Check default values")

	assert.Regexp(t, `[\d]{4}-[\d]{2}-[\d]{2}T[\d]{2}:[\d]{2}:[\d]{2}[.][\d]{3}[+][\d]{2}:[\d]{2}`, logger.Msg.TransactionTime)
}

func capture(fx func(c *fiber.Ctx), ctx *fiber.Ctx) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	fx(ctx)
	log.SetOutput(os.Stderr)
	return buf.String()
}

func TestPrint(t *testing.T) {
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	logger := Logger{}
	logger.Intialize()
	now := time.Now()
	logger.StartTime = now.Add(time.Duration(-5) * time.Second)

	ctx.Locals("logger", &logger)
	out := capture(logger.Print, ctx)

	var result LoggingMessage
	err := json.Unmarshal([]byte(out), &result)
	assert.Equal(t, nil, err, "Well json format")
	assert.Equal(t, uint64(5000), result.ExecutionTimeMsec, "Check execution time calculation: "+out)
}
