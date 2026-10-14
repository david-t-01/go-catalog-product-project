# Build
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Run
FROM alpine:3.21
RUN adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8484
HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD wget -qO- http://localhost:8484/healthz || exit 1
ENTRYPOINT ["/app/server"]
