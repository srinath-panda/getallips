CWD := $(shell pwd)

build:
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/getallips-darwin-amd64
	@env GOOS=darwin GOARCH=arm64 go build -o ./bin/getallips-darwin-arm64
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/getallips-linux-amd64

