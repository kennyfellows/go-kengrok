.PHONY: all
all: cli proxyserver

cli:
	go build -o bin/kengrok cmd/cli/main.go

proxyserver:
	CGO_ENABLED=1 go build -o bin/proxyserver cmd/proxyserver/main.go
