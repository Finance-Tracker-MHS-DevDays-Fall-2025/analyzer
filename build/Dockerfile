FROM golang:1.23-alpine AS build

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/analyzer/ ./cmd
COPY internal ./internal
COPY pkg/ ./pkg
COPY config.yaml ./

RUN GOOS=linux GOARCH=amd64 go build -o ./bin/main ./cmd

FROM alpine:latest AS app

WORKDIR /app

COPY --from=build /app/bin/main ./

CMD ["./main"]
