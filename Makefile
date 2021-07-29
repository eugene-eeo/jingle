build:
	go build

tests:
	gotest ./lexer/
	gotest ./parser/

cover:
	gotest $(arg1) -coverprofile=coverage.out
	go tool cover -func=coverage.out
