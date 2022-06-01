.PHONY: build
build:
	go build -v ./cmd/apiserver
	go build -v ./cmd/rssparser

clean:
	rm apiserver rssparser

.DEFAULT_GOAL := build