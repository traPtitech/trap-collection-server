# syntax = docker/dockerfile:1.0-experimental

FROM groovy:3.0.2 AS generate
WORKDIR /home/groovy/scripts
COPY --chown=groovy:groovy ./docs/swagger/openapi.yml ./generate /local/
VOLUME /home/groovy/.groovy/grapes
RUN groovy /local/generator.groovy generate \
  -i /local/openapi.yml \
  -g CollectionCodegen \
  -t /local \
  -o /local
COPY . /local


FROM golang:1.14.2-alpine AS build

RUN --mount=type=cache,target=/var/cache/apk apk add --update git upx

WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY go.* ./
RUN go mod download

COPY --from=generate /local ./
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -o main -ldflags "-s -w"
RUN upx ./main


FROM alpine:3.11.6 AS runtime

RUN apk --update add tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /Tokyo && \
  apk del tzdata && \
  rm -rf /var/cache/apk/* && \
  mkdir -p /usr/share/zoneinfo/Asia && \
  mv /Tokyo /usr/share/zoneinfo/Asia

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/main ./

ENTRYPOINT ./main