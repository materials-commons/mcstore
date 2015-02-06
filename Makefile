.PHONY: bin test all fmt deploy docs server libs

all: fmt bin

bin: server

server:
	(cd ./server/mcstored; godep go build mcstored.go)

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

libs:
	-godep go install ./...

deploy: server
	-cp server/mcstored/mcstored $$GOPATH/bin
