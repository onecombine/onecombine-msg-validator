package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]

	mode, _ := strconv.Atoi(args[0])
	url := args[1]
	key := args[2]

	switch mode {
	case 0:
		{
			client := &http.Client{}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Liquid-Api-Key", key)
			res, _ := client.Do(req)
			log.Printf("%v\n", res)
		}

	}
}

/*
func GenerateSignature(ctx *fiber.Ctx) error {
	req := new(SignatureRequest)
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("")
	}

	algo := algorithms.NewOneCombineHmac(req.Secret, 60*60*1000)
	hmac := algo.(*algorithms.OneCombineHmac)
	sig := hmac.Sign(req.Body)

	var resp SignatureResponse
	resp.Signature = sig
	raw, _ := json.Marshal(resp)
	return ctx.Status(fiber.StatusOK).SendString(string(raw))
}
*/
