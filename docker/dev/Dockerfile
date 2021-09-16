# syntax = docker/dockerfile:1.0-experimental

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

FROM golang:1.17.1-alpine AS build

RUN apk --update --no-cache add tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

ENV DOCKERIZE_VERSION v0.6.1

RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

RUN apk add --update --no-cache git

ENV GO111MODULE=on

RUN --mount=type=cache,target=/root/.cache/go-build \
  GO111MODULE=off go get github.com/oxequa/realize && \
  rm -rf /go/src/*

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

RUN mkdir upload

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
  go mod download

COPY --from=generate /local/openapi ./openapi
COPY .realize.yaml ./

ENTRYPOINT dockerize -wait tcp://mariadb:3306 realize start --name='server' --install --run
