package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/onecombine/onecombine-msg-validator/src/algorithms"
)

func main() {
	args := os.Args[1:]

	mode, _ := strconv.Atoi(args[0])
	url := args[1]
	apikey := args[2]
	secret := args[3]

	switch mode {
	case 0:
		{
			client := &http.Client{}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Liquid-Api-Key", apikey)
			res, err := client.Do(req)

			if err == nil {
				defer res.Body.Close()

				if res.StatusCode < 300 {
					bytes, err := io.ReadAll(res.Body)
					if err != nil {
						log.Fatal(err)
					}
					body := string(bytes)
					log.Printf("%v\n", body)
				}
			}
		}
	case 1:
		{
			body := `{"partner_id": "500001", "order_ref": "2020040318061601678097480", "payee": "400001100000000", "currency_code": "SGD", "amount": "8.80", "service_code": "13", "merchant_name": "O Coffee Club", "merchant_city": "Singapore", "merchant_country_code": "SG", "mcc": "5812", "postal_code": "138577", "payload_code": "XNAP"}`
			algo := algorithms.NewOneCombineHmac(secret, 60*60*1000)
			hmac := algo.(*algorithms.OneCombineHmac)
			sig := hmac.Sign(body)
			log.Printf("sig %v\n", sig)

			client := &http.Client{}
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
			req.Header.Set("Liquid-Api-Key", apikey)
			req.Header.Set("Signature", sig)
			req.Header.Set("Content-Type", "application/json")
			res, err := client.Do(req)

			if err == nil {
				defer res.Body.Close()

				if res.StatusCode < 300 {
					bytes, err := io.ReadAll(res.Body)
					if err != nil {
						log.Fatal(err)
					}
					body := string(bytes)
					log.Printf("%v\n", body)
				}
			}
		}
	case 2:
		{
			client := &http.Client{}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Liquid-Api-Key", apikey)
			req.Header.Set("Content-Type", "application/json")
			res, err := client.Do(req)

			if err == nil {
				defer res.Body.Close()

				if res.StatusCode < 300 {
					bytes, err := io.ReadAll(res.Body)
					if err != nil {
						log.Fatal(err)
					}
					body := string(bytes)
					log.Printf("%v\n", body)
				}
			}
		}
	case 99:
		{
			body := `{"partner_id": "500001", "order_ref": "2020040318061601678097480", "payee": "400001100000000", "currency_code": "SGD", "amount": "8.80", "service_code": "13", "merchant_name": "O'\'' Coffee Club", "merchant_city": "Singapore", "merchant_country_code": "SG", "mcc": "5812", "postal_code": "138577", "payload_code": "XNAP"}`
			algo := algorithms.NewOneCombineHmac(secret, 60*60*1000)
			hmac := algo.(*algorithms.OneCombineHmac)
			sig := hmac.Sign(body)
			log.Printf("sig %v\n", sig)

			result := hmac.Verify([]byte(body), sig)
			log.Printf("Result %v\n", result)
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
