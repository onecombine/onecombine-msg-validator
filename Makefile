.PHONY: init update run build test clean

init:
	rm -f go.mod
	go mod init github.com/onecombine/onecombine-msg-validator
	go get github.com/gofiber/fiber/v2
	go get github.com/gofiber/fiber/v2/middleware/logger
	go get github.com/aws/aws-sdk-go-v2/aws
	go get github.com/aws/aws-sdk-go-v2/config
	go get github.com/aws/aws-sdk-go-v2/credentials
	go get github.com/aws/aws-sdk-go-v2/service/secretsmanager
	go get github.com/redis/go-redis/v9
	go get github.com/segmentio/kafka-go
	go get github.com/stretchr/testify

run_00:
	go run . 0 "http://localhost:3000/api/v1/qr" "ABCD-ABCD-ABCD" "aaaa"

run_01:
	go run . 1 "http://localhost:3000/api/v1/qr" "ABCD-ABCD-ABCD" "aaaa"

run_02:
	go run . 2 "http://localhost:3000/api/v1/status/33333" "ABCD-ABCD-ABCD" "aaaa"

run_99:
	go run . 99 "http://localhost:3000/api/v1/qr" "ABCD-ABCD-ABCD" "aaaa"

update:
	go get .
	go mod vendor

test:
	go test -v -cover ./...

clean:
	go mod tidy -v
	go mod vendor

publish:
	git tag `cat version`
	git push origin `cat version`
