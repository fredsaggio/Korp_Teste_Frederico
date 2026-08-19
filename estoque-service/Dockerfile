ARG GO_VERSION=1.26.0

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o estoque-service ./cmd

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=build --chown=app:app /app/estoque-service .

USER app

EXPOSE 5001

CMD ["./estoque-service"]
