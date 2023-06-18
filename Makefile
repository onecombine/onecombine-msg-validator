.PHONY: init update run build test clean

init:
	rm -f go.mod
	go mod init github.com/onecombine/onecombine-msg-validator
	go get github.com/aws/aws-sdk-go-v2/aws
	go get github.com/aws/aws-sdk-go-v2/config
	go get github.com/aws/aws-sdk-go-v2/service/secretsmanager

update:
	go get .
	go mod vendor

test:
	go test github.com/onecombine/onecombine-msg-validator/src/algorithms

clean:
	go mod tidy -v
	go mod vendor

publish:
	git tag `cat version`
	git push origin `cat version`