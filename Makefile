.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o ./txParser cmd/main.go
	
.PHONY: test
test:
	@echo "Running tests..."
	go test ./... -v 

.PHONY: run
run:
	go run main.go

.PHONY: dep
dep:
	@echo "Updating dependencies..."
	go mod download

.PHONY: tidy
tidy:
	@echo "go mod tidy..."
	go mod tidy

.PHONY: fmt
fmt:
	@echo "gofmt..."
	gofmt -w -s $$(find . -type f -name '*.go' | grep -v /vendor/)

.PHONY: lint
lint:
	@echo "golangci-lint..."
	golangci-lint run

.PHONY: vet
vet:
	@echo "go vet..."
	go vet ./...

.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -f ./txParser
