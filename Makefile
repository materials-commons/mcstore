.PHONY: bin test all fmt deploy docs server libs cli mc mcbulk

all: fmt bin

bin: server cli

server:
	(cd ./server/mcstore/main; godep go build mcstored.go)

cli: mc mcbulk

mc:
	(cd ./cmd/mccli/main; godep go build mc.go)

mcbulk:
	(cd ./server/cmd/mcbulk; godep go build mcbulk.go)

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

libs:
	-godep go install ./...

deploy: server
	-cp server/mcstore/main/mcstored $$GOPATH/bin
