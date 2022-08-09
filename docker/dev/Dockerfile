# syntax = docker/dockerfile:1.4.1

FROM golang:1.19.0-alpine AS build

RUN apk --update --no-cache add tzdata git && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

ENV AIR_VERSION v1.40.4
RUN --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/cosmtrek/air@$AIR_VERSION

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN go mod download

ENTRYPOINT ["air"]
CMD ["-c", ".air.toml"]
