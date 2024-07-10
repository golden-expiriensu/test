GOBUILD=CGO_ENABLED=0 go build

build:
	@cd cmd/server && $(GOBUILD)

build-for-docker:
	@cd cmd/server && $(GOBUILD) -o /streams-server

start:
	@docker compose down && docker compose up --build

stop:
	@docker compose down

restart: stop start

test:
	@go test ./...

tidy:
	@go mod tidy

lint:
	@gofmt -w -s . && go mod tidy && go vet ./... && golangci-lint run
