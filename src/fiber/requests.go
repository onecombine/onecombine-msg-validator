package fiber

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
	"github.com/onecombine/onecombine-msg-validator/src/utils"
)

const MESSAGE_EXPIRATION_MSEC string = "MESSAGE_EXPIRATION_MSEC"

type XnapUtility struct {
	ApiKey    string
	Validator *algorithms.Validator
}

type Config struct {
	ErrorHandler fiber.Handler
	ApiKeys      map[string]*algorithms.Validator
	Xnap         XnapUtility
	Name         string
}

func GetAcquirerApiKey(ctx *fiber.Ctx) string {
	return ctx.GetReqHeaders()["Liquid-Api-Key"]
}

func NewConfig(name string) *Config {
	var config Config
	config.ApiKeys = make(map[string]*algorithms.Validator)

	aws := utils.NewAwsUtils()
	apiKeys := aws.GetApiKeysMap()
	exp := utils.GetEnv(MESSAGE_EXPIRATION_MSEC, "600000")
	age, _ := strconv.Atoi(exp)

	for key, val := range apiKeys {
		validator := (algorithms.NewOneCombineHmac(val.SecretKey, int32(age))).(algorithms.Validator)
		config.ApiKeys[key] = &validator
	}
	config.ErrorHandler = nil

	config.Xnap.ApiKey = aws.SecretValues.XnapApiKey
	xnapVal := (algorithms.NewOneCombineHmac(aws.SecretValues.XnapSecretKey, int32(age))).(algorithms.Validator)
	config.Xnap.Validator = &xnapVal
	config.Name = name
	return &config
}

func NewHandler(config Config) fiber.Handler {

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fiber.Ctx) error {
			logger := ctx.Locals("Logger").(utils.Logger)
			logger.Msg.HttpStatus = utils.LOGGING_HTTPSTATUS_UNAUTHORIZED
			logger.Msg.ErrorType = utils.LOGGING_ERRORTYPE_BUSINESSERROR
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
		deviceId := ctx.GetReqHeaders()["X-Device-ID"]
		if deviceId != "" {
			opts = append(opts, utils.WithDeviceId(deviceId))
		}
		userAgent := ctx.GetReqHeaders()["User-Agent"]
		if userAgent != "" {
			opts = append(opts, utils.WithUserAgent(userAgent))
		}
		ip := ctx.IP()
		if ip != "" {
			opts = append(opts, utils.WithRemoteAddress(ip))
		}
		userId := ctx.GetReqHeaders()["X-User-ID"]
		if userId != "" {
			opts = append(opts, utils.WithUserId(userId))
		}
		channelId := ctx.GetReqHeaders()["X-Channel-ID"]
		if channelId != "" {
			opts = append(opts, utils.WithChannelId(channelId))
		}
		if config.Name != "" {
			opts = append(opts, utils.WithService(config.Name))
		}
		opts = append(opts, utils.WithRawUrl(ctx.OriginalURL()))
		opts = append(opts, utils.WithHttpMethod(ctx.Method()))

		logger.Intialize(opts...)
		ctx.Locals("Logger", logger)
		defer logger.Print()

		apiKey := ctx.GetReqHeaders()["Liquid-Api-Key"]

		switch ctx.Method() {
		case "GET":
			if config.ApiKeys[apiKey] == nil {
				return config.ErrorHandler(ctx)
			} else {
				return ctx.Next()
			}
		case "POST":
			fallthrough
		case "PUT":
			fallthrough
		case "DELETE":
			validator := config.ApiKeys[apiKey]
			if validator == nil {
				return config.ErrorHandler(ctx)
			} else {
				signature := ctx.GetReqHeaders()["Signature"]
				if (*validator).Verify(ctx.Body(), signature) {
					return ctx.Next()
				} else {
					logger.Msg.HttpStatus = utils.LOGGING_HTTPSTATUS_UNAUTHORIZED
					logger.Msg.ErrorType = utils.LOGGING_ERRORTYPE_BUSINESSERROR
					return ctx.SendStatus(fiber.StatusUnauthorized)
				}
			}
		default:
			return config.ErrorHandler(ctx)
		}
	}
}
