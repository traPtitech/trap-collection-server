# syntax = docker/dockerfile:1

FROM golang:1.26.5-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build

RUN apk add --update --no-cache git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY ./ ./

RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go generate ./... \
  && go build -o collection -ldflags "-s -w"

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

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
