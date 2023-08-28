package fiber

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
	"github.com/onecombine/onecombine-msg-validator/src/utils"
)

type Config struct {
	ErrorHandler fiber.Handler
	ApiKeys      map[string]*AcquirerUtility
	Xnap         XnapUtility
	Name         string
}

func GetAcquirerApiKey(ctx *fiber.Ctx) string {
	return ctx.GetReqHeaders()["Liquid-Api-Key"]
}

func NewConfig(name string) *Config {
	var config Config
	config.ApiKeys = make(map[string]*AcquirerUtility)

	aws := utils.NewAwsSecretValues(nil)
	apiKeys := aws.GetApiKeysMap()
	exp := utils.GetEnv(MESSAGE_EXPIRATION_MSEC, "600000")
	age, _ := strconv.Atoi(exp)

	for key, val := range apiKeys {
		validator := (algorithms.NewOneCombineHmac(val.SecretKey, int32(age))).(algorithms.Validator)
		config.ApiKeys[key] = &AcquirerUtility{validator: &validator, id: val.Id}
	}
	config.ErrorHandler = nil

	config.Xnap.ApiKey = aws.XnapApiKey
	xnapVal := (algorithms.NewOneCombineHmac(aws.XnapSecretKey, int32(age))).(algorithms.Validator)
	config.Xnap.Validator = &xnapVal
	config.Name = name
	return &config
}

func NewHandler(config Config) fiber.Handler {
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
		apiKey := ctx.GetReqHeaders()["Liquid-Api-Key"]
		acquirer := config.ApiKeys[apiKey]

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
		} else if acquirer != nil {
			opts = append(opts, utils.WithPartnerId(acquirer.id))
		}

		if config.Name != "" {
			opts = append(opts, utils.WithService(config.Name))
		}
		opts = append(opts, utils.WithRawUrl(ctx.BaseURL()+ctx.OriginalURL()))
		opts = append(opts, utils.WithHttpMethod(ctx.Method()))

		logger.Intialize(opts...)
		ctx.Locals("logger", &logger)

		switch ctx.Method() {
		case "GET":
			if acquirer == nil {
				err := config.ErrorHandler(ctx)
				defer logger.Print(ctx)
				return err
			} else {
				err := ctx.Next()
				defer logger.Print(ctx)
				return err
			}
		case "POST":
			fallthrough
		case "PUT":
			fallthrough
		case "DELETE":
			validator := acquirer.validator
			if validator == nil {
				err := config.ErrorHandler(ctx)
				defer logger.Print(ctx)
				return err
			} else {
				signature := ctx.GetReqHeaders()["Signature"]
				if (*validator).Verify(ctx.Body(), signature) {
					err := ctx.Next()
					defer logger.Print(ctx)
					return err
				} else {
					logger.Msg.HttpStatus = utils.LOGGING_HTTPSTATUS_UNAUTHORIZED
					logger.Msg.ErrorType = utils.LOGGING_ERRORTYPE_BUSINESSERROR
					err := ctx.SendStatus(fiber.StatusUnauthorized)
					defer logger.Print(ctx)
					return err
				}
			}
		default:
			err := config.ErrorHandler(ctx)
			defer logger.Print(ctx)
			return err
		}
	}
}
