.PHONY: bin test all fmt deploy docs server libs cli

all: fmt bin

bin: server cli

server:
	(cd ./server/mcstore/main; godep go build mcstored.go)

cli:
	(cd ./cmd/mc; godep go build mc.go)

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

libs:
	-godep go install ./...

deploy: server
	-cp server/mcstored/mcstored $$GOPATH/bin
