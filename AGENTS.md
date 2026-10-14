# AGENTS.md: how I work on Go projects, and the rules of this one

Vendor-neutral instructions for AI agents (and humans). Part A is my way of building Go
services, learned from my previous projects (Gin REST APIs and Redis-driven workers around a
Kotlin/Node.js/Symfony platform, Postgres and MongoDB storage, docker compose on one host) and updated
to current Go. Part B is how it applies to this repository. When a rule in Part B is more
specific, it wins here; when in doubt, keep the implementation as simple as possible and
idiomatic.

---

## Part A: my style

### A1. Shape of a service

- `cmd/<binary>/main.go` plus one package per role: config, constants (`defines`), domain
  (entities, DTOs, errors, no I/O), service (rules), repository (one per store), transport
  (HTTP handlers or a worker loop), metrics. Migrations, docs, Dockerfile, Makefile and lint
  config at the root.
- One file per entity per layer, snake_case. Binaries named by role: `-api`, `-job`, `-task`. Binary names say what they are: `<thing>-api` (HTTP), `<thing>-job` (processor),
  `<thing>-task` (forwarder).
- `main` is tiny: config, logger, composition root, run. Wiring happens in one function, in
  order, under section comments (`// Clients`, `// Repositories`, `// Services`,
  `// Handlers`). Workers have no controller layer.
- Start every service with a scaffold (folders, `main`, README skeleton), then the logic.

### A2. Layers and interfaces

| Layer | Owns | Never |
| --- | --- | --- |
| handler / worker loop | parsing input, calling the service, mapping errors to transport | business rules, SQL, constructing dependencies |
| service | rules: defaults, clamping, validation, orchestration, mapping to DTOs | transport types (`*gin.Context`, `http.Request`), drivers |
| repository | queries, driver error translation to sentinels | business rules, logging |
| domain / defines | entities, DTOs, errors, constants | I/O |

- Interfaces are declared where they are consumed and sized to what the consumer needs (one to
  three methods); the service declares the repository interface, the handler declares the
  service interface. Constructors return the concrete `*Type`. No interface without a second
  implementation or a test double.
- No pass-through layers: if a service method only forwards to the repository, the layer is not
  earning its place.
- Extension by embedding (a shared struct or interface inside a local one) is fine; embedding an
  interface as a nil field to borrow its method set is not.

### A3. Naming

- Packages short, singular, named by what they provide. Constants grouped by family with a
  prefix (`Env*`, `Endpoint*`, `Queue*`) in a constants package.
- Receivers are short role words (`svc`, `repo`, `h`); acronyms keep their case (`ID`, `SKU`);
  no `package.PackageX` stutter.
- Request DTOs end in `Request` and carry `Validate()`; response DTOs use snake_case JSON tags
  and never expose database ids.
- Environment variables are `UPPER_SNAKE_CASE` with a backing-service prefix (`REDIS_*`,
  `POSTGRES_*`), topology as configuration (`REDIS_CHANNEL_*`, `*_Q`), flags `FEATURE_FLAG_*`.
- Code and comments in English; user-facing strings localised and kept in one place.

### A4. Configuration and errors

- 12-factor: everything from the environment, loaded once at the root into a typed struct that
  validates and fails fast. Nothing below the root reads the environment; `.env` is a
  convenience for host runs, never required.
- Errors are wrapped with `%w` and an operation prefix, checked with `errors.Is/As`. Sentinels
  for conditions callers branch on (`ErrNotFound`, `ErrConflict`) and a typed validation error;
  never `(nil, nil)` for not found.
- The transport layer maps errors to HTTP in one place per package; internal details are
  logged, never returned. Handle an error once: log it or return it.
- Batch jobs collect per-row errors and continue; only infrastructure errors abort.

### A5. Context, lifecycle, concurrency

- `ctx` is the first parameter of everything that does I/O; never stored, never created below
  the root.
- `main` uses `signal.NotifyContext` and delegates to `run() error` so every defer runs; HTTP
  servers have timeouts and a bounded `Shutdown`; every client is closed on exit.
- Workers subscribe once, bound in-flight work, own their goroutines with `errgroup`, and clean
  up registries and dedup keys on every exit path. HTTP clients always have a timeout.

### A6. Data and messaging

- Explicit queries with context, explicit columns, closed rows, non-nil empty slices,
  transactions where two writes must agree. Migrations are numbered files run in order; never
  edited after they land.
- Money is `decimal`; `float64` only in response DTOs.
- Redis is the bus between services (`Publish` for events, lists for queues, keys with TTL for
  caches), names from configuration.

### A7. Logging and metrics

- `log/slog`, injected, never global; structured attributes; log at the edge with a request or
  job id, inner layers return errors. Keep a platform's log line format when something tails
  the files.
- Metrics through middleware and instrumented clients, never counters by hand in handlers.

### A8. Testing (TDD) and linting

- Red, green, refactor: failing test first, run it, minimum implementation, green, clean up.
- Table-driven unit tests next to the code; handlers with `httptest`; services with mocks of the
  consumer interfaces (mockery, black-box `_test` packages); repositories with integration
  tests against a real database behind a build tag.
- `go test -race`, `go vet` and golangci-lint with a committed config on every change; fix
  findings, exclude a rule only with a comment saying why. Tools pinned with `go tool`.

### A9. Build and operations

- Two-stage Docker image: pinned builder, cached module download, static binary, pinned minimal
  runtime, non-root user, healthcheck, exec-form entrypoint, no entrypoint script.
- Compose starts the database by default and the app behind a profile; healthchecks and
  `condition: service_healthy` everywhere; shared plus per-service env files with committed
  examples; pinned images.
- A shared library is a tagged module, never a `replace` in a committed `go.mod`.

### A10. Docs, git, comments

- README: purpose, Idea/Concept, Steps/Flow, endpoints, configuration, run and test, decisions,
  useful commands. Diagrams as text next to their image; an OpenAPI spec served by the binary;
  a Postman collection for manual checks.
- Commits `feat: [TASK-NNN] - brief description`, one branch per task from `main`, every branch
  ends runnable. The human owns git: agents never commit, push, or touch branches; they stop at
  the commit point, list the files and propose the message.
- Comments only for large or non-obvious logic (why, not what), at most 5-10 lines, never on
  self-explanatory code. No commented-out code, no dead files.

### A11. Do not

- No interface next to its only implementation, no interfaces returned from constructors.
- No environment reads outside the root; no transport types below the handler.
- No `errors.New(fmt.Sprintf(...))`, no Yoda conditions, no `err1`/`err2`, no `*[]T`.
- No `context.Background()` in repositories; no zero-value `http.Client`; no subscribe inside a
  loop; no unbounded goroutines; nothing left unclosed.
- No root containers, shell-form entrypoints, floating image tags, or committed `.env` values.

---

## Part B: this project

### B1. What it is

A products catalog REST API: categories, catalog listing with offset pagination and filters,
product details with variants, category creation. Postgres through GORM, standard library
`net/http` ServeMux, requirements in `ASSIGNMENT.md`, interactive docs at `/docs/`. Small by
design: no framework, no config package, no code generation.

### B2. Commands

- `make docker-up` (database), `make docker-up-app` (database plus API container),
  `make seed` (recreates all tables from `sql/`), `make run`.
- `make test`, `make test-integration` (needs the database), `make lint`, `make mocks`,
  `make docs-ui`, `make docker-build`.
- Before every hand-off: `make test && make lint && go vet ./...`; integration tests after
  repository changes; a container smoke test after Docker or compose changes.

### B3. Layout and contract

- `app/api` writes every JSON response; `app/apidocs` embeds the spec and UI; `app/<feature>`
  has `handler.go`, `service.go`, `dto.go` and generated `mocks/`; `models` has the GORM models,
  sentinel errors and repositories; `sql/` the numbered migrations; `postman/` the collection.
- Consumer interfaces: `catalog.ProductsRepository` and `categories.Repository` (services),
  `catalog.Catalog` and `categories.Categories` (handlers). Concrete `Service` and `Handler`.
- The HTTP contract (paths, pagination defaults and clamping, exact error bodies) is the spec in
  `app/apidocs/static/openapi.yaml`. Changing an endpoint means changing the spec, the tests and
  the Postman collection together.

### B4. Testing here

- Unit tests in `<feature>_test` packages with mocks from `.mockery.yml`; regenerate with
  `make mocks` when an interface changes. Repository tests are integration tests on the seeded
  database and leave the seed as they found it. `app/api/response_test.go` is never edited.

### B5. Workflow for new requirements

Use the superpowers plugin skills, in order, stopping at each gate for the human:

1. `superpowers:brainstorming`: classify the request, write back the understanding, one
   question at a time, two or three approaches, design in sections, approval before building.
2. `superpowers:writing-plans` for architectural work: bite-sized tasks, each with its failing
   test, implementation and verification.
3. `superpowers:executing-plans` (or `subagent-driven-development` when asked), with
   `superpowers:test-driven-development` inside every task.
4. `superpowers:verification-before-completion`: run B2 and report the real output.
5. `/simplify` and the `code-simplifier` agent over the changed files, preserving behaviour and
   the contract; then B2 again.
6. Bugs: `superpowers:systematic-debugging` before any fix.

Library choices go to a read-only research subagent that returns a comparison and a
recommendation; the human picks.

### B6. Git here

- Branches `feature/gocat-NNN-<slug>` created by the human from `main`; messages
  `feat: [TASK-NNN] - brief description`; the human commits and merges. Agents use read-only
  git only and pass the same rule to subagents.
