.PHONY: bin test all fmt deploy docs test-client test-server test-base bin-client bin-server libs

all: fmt test bin

bin: bin-client bin-server

bin-client:
	(cd ./client/main; godep go build materials.go)

bin-server:
	(cd ./server/main; godep go build mcfs.go)

test: test-client test-server test-base

test-client:
	(cd ./client; make test)

test-server:
	(cd ./server; make test)

test-base:
	(cd ./base; make test)

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

libs:
	-godep go install ./...

deploy: test-server bin-server
	-cp server/main/mcfs $$GOPATH/bin
