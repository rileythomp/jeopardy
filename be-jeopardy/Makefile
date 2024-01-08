BINARY_NAME=bin/jeopardy
GOBUILD=go build
GOCLEAN=go clean


build:
	go build -o $(BINARY_NAME) -v

clean:
	go clean
	rm -f $(BINARY_NAME)

run:
	go build -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

all: build

.PHONY: all build clean run
