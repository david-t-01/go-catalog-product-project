# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

1. **cmd/**: Contains the main application and seed command entry points.

   - `server/main.go`: The main application entry point, serves the REST API.
   - `seed/main.go`: Command to seed the database with initial product data.

2. **app/**: Contains the application logic.
3. **sql/**: Contains a very simple database migration scripts setup.
4. **models/**: Contains the data models and repositories used in the application.
5. `.env`: Environment variables file for configuration.

## Setup Code Repository

1. Create a github/bitbucket/gitlab repository and push all this code as-is.
2. Create a new branch, and provide a pull-request against the main branch with your changes. Instructions to follow.

## Application Setup

- Ensure you have Go installed on your machine.
- Ensure you have Docker installed on your machine.
- Important makefile targets:
  - `make tidy`: will install all dependencies.
  - `make docker-up`: will start the required infrastructure services via docker containers.
  - `make seed`: ⚠️ Will destroy and re-create the database tables.
  - `make test`: Will run the tests.
  - `make run`: Will start the application.
  - `make docker-down`: Will stop the docker containers.

Follow up for the assignemnt here: [ASSIGNMENT.md](ASSIGNMENT.md)

## Useful commands

Infrastructure

- `make docker-up` / `make docker-down`: start or stop Postgres (only the database; the API runs on the host with `make run`).
- `docker compose logs -f postgres`: follow the database logs.
- `docker compose exec postgres psql -U postgres -d challenge`: open a SQL shell in the container (`\dt` lists tables, `\q` quits).

Database seed

- `make seed`: runs every `sql/*.sql` in name order from the host against `localhost:5432`. Destructive: drops and recreates all tables.
- Seed through the container, piping each file into `psql` (no Go needed on the host):

  ```bash
  for f in sql/*.sql; do docker compose exec -T postgres psql -q -U postgres -d challenge -v ON_ERROR_STOP=1 < "$f"; done
  ```

- Seed with the Go command inside a throwaway container on the compose network (uses the same `cmd/seed`):

  ```bash
  docker run --rm --network go-catalog-product-project_default -v "$PWD":/src -w /src \
    -e POSTGRES_HOST=postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password \
    -e POSTGRES_DB=challenge -e POSTGRES_PORT=5432 -e POSTGRES_SQL_DIR=./sql \
    golang:1.26-alpine go run ./cmd/seed
  ```

- Quick check: `docker compose exec -T postgres psql -U postgres -d challenge -tAc "select count(*) from products;"` prints `8`.

## Build, test, lint

- `make run`: start the API on `http://localhost:8484`.
- `make test`: unit tests with the race detector and coverage (no Docker needed).
- `make test-integration`: repository tests against the compose database (`make docker-up && make seed` first).
- `make lint`: golangci-lint v2 (pinned as a Go tool: `go tool golangci-lint version`).
- `make mocks`: regenerate mocks from `.mockery.yml` (`go tool mockery version`).
- `go vet ./...`, `go tool golangci-lint fmt ./...`: vet and format.

### Manual API checks

- Import `postman/go-catalog.postman_collection.json` and `postman/local.postman_environment.json` into Postman; one folder per branch, every request asserts status and body.
- `curl -s localhost:8484/catalog | jq .` for a quick look from the terminal.

## Endpoints

| Method | Path | Query / body | Responses |
| --- | --- | --- | --- |
| GET | `/healthz` | - | `200 {"status":"ok"}` |
| GET | `/catalog` | `offset` (default 0, negative clamped to 0), `limit` (default 10, clamped to 1..100), `category=<CODE>` (exact), `price_less_than=<decimal>` (strict) | `200 {"products":[{"code","price","category":{"code","name"}}],"total","offset","limit"}`; `400 {"error":"invalid limit: must be an integer"}` |
| GET | `/catalog/{code}` | - | `200 {"code","price","category","variants":[{"name","sku","price"}]}` (variants without a price inherit the product price); `404 {"error":"product not found"}` |
| GET | `/categories` | - | `200 {"categories":[{"code","name"}]}` |
| POST | `/categories` | `{"code":"BAGS","name":"Bags"}` (trimmed; code <= 32, name <= 256) | `201 {"code","name"}`; `400 {"error":"code is required"}` / `{"error":"invalid JSON body"}`; `409 {"error":"category code already exists"}` |

Examples: `curl 'localhost:8484/catalog?category=SHOES&limit=5'`, `curl -X POST localhost:8484/categories -d '{"code":"BAGS","name":"Bags"}'`.

## API docs

Interactive OpenAPI documentation is served by the API itself at `http://localhost:8484/docs/` (spec at `/docs/openapi.yaml`, source in `app/apidocs/static/openapi.yaml`). The Swagger UI assets are vendored; refresh them with `make docs-ui`.

## Decisions

- Three layers per feature (`handler -> service -> repository`). Interfaces are declared by their consumers (`catalog.Catalog`, `catalog.ProductsRepository`, `categories.Categories`, `categories.Repository`) and constructors return concrete types.
- Handlers parse strings and map errors to HTTP once; services own the rules (pagination defaults and clamping, trimming, validation, variant price inheritance); repositories own the GORM queries.
- Out-of-range `offset`/`limit` are clamped; non-numeric values are rejected with 400. The category filter matches the code exactly (case-sensitive).
- Category `id` is never exposed. Variant prices are `decimal.NullDecimal`: NULL inherits the product price, an explicit `0.00` is kept.
- Errors are sentinels (`ErrNotFound`, `ErrConflict`, `ValidationError`) wrapped with `%w`; GORM `TranslateError` turns duplicate keys into `ErrConflict`.
- Unit tests use mockery v3 mocks in black-box `_test` packages; repositories are tested with `make test-integration` against the compose database.
- The API image is opt-in: `make docker-up` starts only Postgres, `make docker-up-app` the full stack. The container reaches the database through `POSTGRES_HOST`.

## Development

- `make lint`: golangci-lint v2 (pinned as a Go tool). `make mocks`: regenerate mocks from `.mockery.yml`. `make test-integration`: repository tests (needs `make docker-up` and `make seed`).
- `make docker-build`, `make docker-up-app`, `make docker-down`, `make docs-ui`.
