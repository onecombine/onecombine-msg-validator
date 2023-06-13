.PHONY: init update run build test clean

init:
	rm -f go.mod
	go mod init github.com/onecombine/onecombine-msg-validator

update:
	go get .
	go mod vendor

test:
	go test github.com/onecombine/onecombine-msg-validator/src

clean:
	go mod tidy -v
	go mod vendor

publish:
	git tag `cat version`
	git push origin `cat version`