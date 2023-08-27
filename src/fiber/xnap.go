package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

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
	store := session.New()

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fiber.Ctx) error {
			session := utils.GetLoggingSession()
			logger := session.Get(ctx)
			logger.Msg.HttpStatus = utils.LOGGING_HTTPSTATUS_UNAUTHORIZED
			logger.Msg.ErrorType = utils.LOGGING_ERRORTYPE_BUSINESSERROR
			session.Save(ctx, logger)
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

		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}
		ctx.Locals("Session-ID", sess.ID())
		session := utils.GetLoggingSession()
		session.Save(ctx, &logger)
		defer session.Flush(ctx)

		switch ctx.Method() {
		case "POST":
			err := ctx.Next()
			session.Save(ctx, logger.Collect(ctx))
			return err
		case "GET":
			fallthrough
		case "PUT":
			fallthrough
		case "DELETE":
			fallthrough
		default:
			err := config.ErrorHandler(ctx)
			session.Save(ctx, logger.Collect(ctx))
			return err
		}
	}
}
