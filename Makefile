.PHONY: build test clean install

build:
	go build -o bin/cr ./cmd/cr

test:
	go test -v ./...

clean:
	rm -rf bin
	rm -rf review_results
	rm -rf .cache

install:
	go install ./cmd/cr 