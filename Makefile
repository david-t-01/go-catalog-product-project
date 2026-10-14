tidy ::
	@go mod tidy && go mod vendor

seed ::
	@go run cmd/seed/main.go

run ::
	@go run cmd/server/main.go

test ::
	@go test -v -count=1 -race ./... -coverprofile=coverage.out -covermode=atomic

docker-up ::
	docker compose up -d

docker-down ::
	docker compose --profile app down

docker-build ::
	docker build -t go-catalog-api .

docker-up-app ::
	docker compose --profile app up --build -d

lint ::
	@go tool golangci-lint run ./...

mocks ::
	@go tool mockery

test-integration ::
	@go test -count=1 -race -tags integration ./models/...
