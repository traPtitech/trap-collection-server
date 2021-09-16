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
COPY . /local


FROM golang:1.17.1-alpine AS build

RUN apk add --update --no-cache git

WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY go.mod go.sum ./
RUN go mod download

COPY --from=generate /local/ ./
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -o main -ldflags "-s -w" -tags main

FROM alpine:3.13.5 AS runtime

ENV TZ=Asia/Tokyo
RUN apk --update --no-cache add tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

ENV DOCKERIZE_VERSION=v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/main ./
COPY ./upload ./upload

ENTRYPOINT dockerize -wait tcp://collection-mariadb:3306 ./main