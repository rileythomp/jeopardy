BINARY_NAME=bin/jeopardy

build:
	go build -o $(BINARY_NAME) -v

run: build
	./$(BINARY_NAME)

clean:
	go clean
	rm -f $(BINARY_NAME)

run-heroku:
	go install -v ./...
	heroku local web --port 8080

.PHONY: build run clean run-heroku
