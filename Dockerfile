# Web Server
FROM golang:1.22.4-alpine AS webserver

RUN apk add --update --no-cache git curl

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o webserver ./example/goroutine-leak/cmd/webserver/main.go
RUN go build -o loadgen ./example/goroutine-leak/cmd/loadgen/main.go

CMD /app/webserver

EXPOSE 8080

# Load Generator AB
FROM alpine:3.20.1 AS ab

RUN apk add --update --no-cache apache2-utils

# Load Generator Golang
FROM alpine:3.20.1 AS loadgen

WORKDIR /app

COPY --from=webserver /app/loadgen .


