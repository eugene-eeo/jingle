default: tests build
build:
	go generate ./scanner/
	go build

tests:
	gotest ./scanner/
	gotest ./parser/

cover:
	gotest $(arg1) -coverprofile=coverage.out
	go tool cover -func=coverage.out
