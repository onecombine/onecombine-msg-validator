package fiber

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
	"github.com/onecombine/onecombine-msg-validator/src/partners"
	"github.com/onecombine/onecombine-msg-validator/src/utils"
)

type Config struct {
	ErrorHandler fiber.Handler
	ApiKeys      map[string]*AcquirerUtility
	Xnap         XnapUtility
	Name         string
}

func GetAcquirerApiKey(ctx *fiber.Ctx) string {
	apiKey, ok := ctx.GetReqHeaders()["X-Api-Key"]
	if !ok {
		apiKey = ctx.GetReqHeaders()["Liquid-Api-Key"]
	}
	return apiKey
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

func NewPartnerConfig(name string, s partners.PartnerService) *Config {
	var config Config
	aws := utils.NewAwsSecretValues(nil)
	config.ApiKeys = make(map[string]*AcquirerUtility)
	acqs, err := s.ListAcquirers()
	if err != nil {
		panic(err)
	}

	exp := utils.GetEnv(MESSAGE_EXPIRATION_MSEC, "600000")
	age, _ := strconv.Atoi(exp)

	for _, a := range acqs {
		validator := (algorithms.NewOneCombineHmac(a.Secret, int32(age))).(algorithms.Validator)
		config.ApiKeys[a.ApiKey] = &AcquirerUtility{validator: &validator, id: a.Name}
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
			err := APIError{
				ErrorCode:        UNAUTHORIZED_ERROR_CODE,
				ErrorDescription: UNAUTHORIZED_ERROR_DESC,
			}

			raw, _ := json.Marshal(err)
			return ctx.Status(fiber.StatusUnauthorized).SendString(string(raw))
		}
	}

	return func(ctx *fiber.Ctx) error {
		apiKey, ok := ctx.GetReqHeaders()["X-Api-Key"]
		if !ok {
			apiKey = ctx.GetReqHeaders()["Liquid-Api-Key"]
		}

		acquirer, ok := config.ApiKeys[apiKey]

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

		// Missing or invalid API-KEY
		if !ok {
			err := config.ErrorHandler(ctx)
			defer logger.Print(ctx)
			return err
		}

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
					errResp := APIError{
						ErrorCode:        INVALID_SIGNATURE_ERROR_CODE,
						ErrorDescription: INVALID_SIGNATURE_ERROR_DESC,
					}
					raw, _ := json.Marshal(errResp)
					err := ctx.Status(fiber.StatusUnauthorized).SendString(string(raw))
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
