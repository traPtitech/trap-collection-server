# syntax = docker/dockerfile:1.3.0

FROM groovy:4.0 AS generate
WORKDIR /home/groovy/scripts
COPY --chown=groovy:groovy ./docs/swagger/openapi.yml ./generate /local/
USER root
RUN --mount=type=cache,target=/home/groovy/.groovy/grapes \
  groovy /local/generator.groovy generate \
  -i /local/openapi.yml \
  -g CollectionCodegen \
  -t /local \
  -o /local

FROM golang:1.18.0-alpine AS build

RUN apk add --update --no-cache git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY --from=generate /local/openapi ./openapi

COPY ./ ./

RUN go generate ./... \
  && go build -o collection -ldflags "-s -w"

FROM alpine:3.15.0

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

RUN apk --update --no-cache add tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && apk del tzdata \
  && mkdir -p /usr/share/zoneinfo/Asia \
  && ln -s /etc/localtime /usr/share/zoneinfo/Asia/Tokyo
RUN apk --update --no-cache add ca-certificates \
  && update-ca-certificates \
  && rm -rf /usr/share/ca-certificates

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/collection ./collection

ENTRYPOINT ./collection
