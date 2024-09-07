# syntax = docker/dockerfile:1

FROM golang:1.23.1-alpine AS build

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add tzdata git && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN go mod download

ENV GOFLAGS -buildvcs=false

ENTRYPOINT ["go", "run", "github.com/cosmtrek/air"]
CMD ["-c", ".air.toml"]
