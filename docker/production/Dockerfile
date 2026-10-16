# syntax = docker/dockerfile:1

FROM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build

RUN apk add --update --no-cache git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY ./ ./

RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go generate ./... \
  && go build -o collection -ldflags "-s -w"

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

RUN apk --update --no-cache add tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && apk del tzdata \
  && mkdir -p /usr/share/zoneinfo/Asia \
  && ln -s /etc/localtime /usr/share/zoneinfo/Asia/Tokyo
RUN apk --update --no-cache add ca-certificates \
  && update-ca-certificates \
  && apk del ca-certificates

COPY --from=arigaio/atlas:1.2.0@sha256:69fef5b506378439771fc70e052471780336552dbf35da397f5a7a729da8f7e3 --chown=root:root atlas /usr/local/bin/atlas

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/collection ./collection

ENTRYPOINT ./collection
