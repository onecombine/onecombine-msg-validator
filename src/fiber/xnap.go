package fiber

import (
	"github.com/gofiber/fiber/v2"

	"github.com/onecombine/onecombine-msg-validator/src/utils"
)

type XnapConfig struct {
	ErrorHandler fiber.Handler
	Name         string
}

func NewXnapConfig(name string) *XnapConfig {
	var config XnapConfig
	config.ErrorHandler = nil
	config.Name = name
	return &config
}

func NewXnapHandler(config XnapConfig) fiber.Handler {
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fiber.Ctx) error {
			logger := ctx.Locals("logger").(*utils.Logger)
			logger.Msg.HttpStatus = utils.LOGGING_HTTPSTATUS_UNAUTHORIZED
			logger.Msg.ErrorType = utils.LOGGING_ERRORTYPE_BUSINESSERROR
			ctx.Locals("logger", logger)
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}
	}

	return func(ctx *fiber.Ctx) error {
		logger := utils.Logger{}

		opts := []utils.Option{}
		reqId := ctx.GetReqHeaders()["X-Request-ID"]
		if reqId != "" {
			opts = append(opts, utils.WithRequestId(reqId))
		}
		userAgent := ctx.GetReqHeaders()["User-Agent"]
		if userAgent != "" {
			opts = append(opts, utils.WithUserAgent(userAgent))
		}
		ip := ctx.IP()
		if ip != "" {
			opts = append(opts, utils.WithRemoteAddress(ip))
		}
		partnerId := ctx.GetReqHeaders()["X-Partner-ID"]
		if partnerId != "" {
			opts = append(opts, utils.WithPartnerId(partnerId))
		}

		if config.Name != "" {
			opts = append(opts, utils.WithService(config.Name))
		}
		opts = append(opts, utils.WithRawUrl(ctx.BaseURL()+ctx.OriginalURL()))
		opts = append(opts, utils.WithHttpMethod(ctx.Method()))

		logger.Intialize(opts...)
		ctx.Locals("logger", logger)

		switch ctx.Method() {
		case "POST":
			err := ctx.Next()
			defer logger.Print(ctx)
			return err
		case "GET":
			fallthrough
		case "PUT":
			fallthrough
		case "DELETE":
			fallthrough
		default:
			err := config.ErrorHandler(ctx)
			defer logger.Print(ctx)
			return err
		}
	}
}
