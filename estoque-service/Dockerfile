FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o estoque-service .

FROM alpine:3.22

WORKDIR /app

COPY --from=build /app/estoque-service .

EXPOSE 5001

CMD ["./estoque-service"]
