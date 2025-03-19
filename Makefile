.PHONY: clean, test, run, dep, tidy, fmt, lint, vet

build:
	go build -o ./txParser main.go
	
test:
	@echo "Running tests..."
	go test ./... -v 

run:
	go run main.go

dep:
	@echo "Updating dependencies..."
	go mod download

tidy:
	@echo "go mod tidy..."
	go mod tidy

fmt:
	@echo "gofmt..."
	gofmt -w -s $$(find . -type f -name '*.go' | grep -v /vendor/)

lint:
	@echo "golangci-lint..."
	golangci-lint run

vet:
	@echo "go vet..."
	go vet ./...

clean:
	@echo "Cleaning up..."
	rm -f ./txParser
