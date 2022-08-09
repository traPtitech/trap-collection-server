# syntax = docker/dockerfile:1.4.1

FROM golang:1.19.0-alpine AS build

RUN apk --update --no-cache add tzdata git build-base && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN go mod download
