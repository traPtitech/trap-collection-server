# syntax = docker/dockerfile:1.0-experimental

FROM golang:1.17.3-alpine AS build

RUN --mount=type=cache,target=/var/cache/apk apk add --update git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY ./mockgen/go.* ./
RUN go mod download

COPY ./mockgen/ ./docs/swagger/openapi.yml ./
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -o main -ldflags "-s -w"
RUN ./main openapi.yml


FROM stoplight/prism:4.6.1 AS main

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/openapi.yml /tmp/