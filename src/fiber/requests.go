package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
)

type Config struct {
	ErrorHandler fiber.Handler
	ApiKeys      map[string]*algorithms.Validator
}

func NewConfig(apiKeys map[string]string, age int32) *Config {
	var config Config
	config.ApiKeys = make(map[string]*algorithms.Validator)
	for key, val := range apiKeys {
		validator := (algorithms.NewOneCombineHmac(val, age)).(algorithms.Validator)
		config.ApiKeys[key] = &validator
	}
	config.ErrorHandler = nil
	return &config
}

func NewHandler(config Config) fiber.Handler {

	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fiber.Ctx) error {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}
	}

	return func(ctx *fiber.Ctx) error {
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
					return ctx.SendStatus(fiber.StatusUnauthorized)
				}
			}
		default:
			return config.ErrorHandler(ctx)
		}
	}
}
