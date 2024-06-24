.DEFAULT_GOAL := app

app:
	docker compose down && docker compose up --build 

lint:
	gofmt -w -s . && go mod tidy && go vet ./...

run-rabbitmq:
	docker run -d --rm --name rabbitmq-test -p 5672:5672 -p 15672:15672 rabbitmq:management

# run rabbitmq before start testing if you don't before
test:
	go clean -testcache && go test ./... -cover -v