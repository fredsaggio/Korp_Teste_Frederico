# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.0
ARG ALPINE_VERSION=3.22

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o faturamento-service ./cmd

FROM alpine:${ALPINE_VERSION}

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=build --chown=app:app /app/faturamento-service .

USER app

EXPOSE 5002

CMD ["./faturamento-service"]
