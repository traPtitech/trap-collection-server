# syntax = docker/dockerfile:1

FROM golang:1.24.5-alpine AS build

RUN apk add --update --no-cache git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY ./ ./

RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go generate ./... \
  && go build -o collection -ldflags "-s -w"

FROM arigaio/atlas:0.35.0 AS atlas

FROM alpine:3.22.1

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

RUN apk --update --no-cache add tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && apk del tzdata \
  && mkdir -p /usr/share/zoneinfo/Asia \
  && ln -s /etc/localtime /usr/share/zoneinfo/Asia/Tokyo
RUN apk --update --no-cache add ca-certificates \
  && update-ca-certificates \
  && apk del ca-certificates

COPY --from=atlas --chown=root:root atlas /usr/local/bin/atlas

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/collection ./collection

ENTRYPOINT ./collection
