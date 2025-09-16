FROM golang:alpine AS builder

RUN apk update && apk upgrade && \
    apk --update add git make bash

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o frontoffice-be main.go

# Distribution
FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    mkdir /app

WORKDIR /app

COPY --from=builder /app /app

CMD /app/frontoffice-be