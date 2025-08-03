# Makefile für Go WebSocket Server

BINARY_NAME=server

.PHONY: build run clean install

build:
	go build -o $(BINARY_NAME) .

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

install:
	go mod tidy
