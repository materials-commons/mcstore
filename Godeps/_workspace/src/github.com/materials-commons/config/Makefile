.PHONY: all test fmt docs install deps

all: fmt test docs

test:
	rm -rf test_data/t
	-godep go test -v ./...

docs:
	./makedocs.sh

fmt:
	-go fmt ./...

install: fmt test
	-godep go install ./...

deps:
	godep save ./...
