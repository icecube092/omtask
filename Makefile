install-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.53.3

install-deps:
	go install github.com/matryer/moq@latest

generate:
	go generate ./...

test:
	go clean -testcache && go test ./...