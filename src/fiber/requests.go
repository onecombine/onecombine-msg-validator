package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
)

type Config struct {
	Checker      algorithms.Validator
	ErrorHandler fiber.Handler
	ApiKeys      map[string]string
}

func NewConfig(key string, age int32) *Config {
	ohmac := algorithms.NewOneCombineHmac(key, age)

	var config Config
	config.Checker = ohmac.(algorithms.Validator)
	config.ErrorHandler = nil
	config.ApiKeys = make(map[string]string)
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
			if config.ApiKeys[apiKey] == "" {
				return config.ErrorHandler(ctx)
			} else {
				return ctx.Next()
			}
		case "POST":
			fallthrough
		case "PUT":
			fallthrough
		case "DELETE":
			if config.ApiKeys[apiKey] == "" {
				return config.ErrorHandler(ctx)
			} else {
				signature := ctx.GetReqHeaders()["Signature"]
				if config.Checker.Verify(ctx.Body(), signature) {
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
