.PHONY: bin test all fmt deploy docs server libs cli

all: fmt bin

bin: server cli

server:
	(cd ./server/mcstore/main; godep go build mcstored.go)

cli:
	(cd ./cmd/mccli/main; godep go build mc.go)
	(cd ./server/cmd/mcbulk; godep go build mcbulk.go)

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

libs:
	-godep go install ./...

deploy: server
	-cp server/mcstore/main/mcstored $$GOPATH/bin
