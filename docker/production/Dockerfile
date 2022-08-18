# syntax = docker/dockerfile:1

FROM golang:1.19.0-alpine AS build

RUN --mount=type=cache,target=/var/cache/apk \
  apk add --update git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY ./ ./

RUN --mount=type=cache,target=/go/pkg/mod/cache \
  --mount=type=cache,target=/root/.cache/go-build \
  go generate ./... \
  && go build -o collection -ldflags "-s -w"

FROM alpine:3.16.2

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && apk del tzdata \
  && mkdir -p /usr/share/zoneinfo/Asia \
  && ln -s /etc/localtime /usr/share/zoneinfo/Asia/Tokyo
RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add ca-certificates \
  && update-ca-certificates \
  && apk del ca-certificates

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/collection ./collection

ENTRYPOINT ./collection
