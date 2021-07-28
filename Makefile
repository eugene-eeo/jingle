build:
	go build

tests:
	gotest ./lexer/
	gotest ./parser/
