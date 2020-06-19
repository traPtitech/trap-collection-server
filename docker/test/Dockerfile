FROM groovy:3.0.2 AS generate
WORKDIR /home/groovy/scripts
COPY --chown=groovy:groovy . /local
VOLUME /home/groovy/.groovy/grapes
RUN groovy /local/generate/generator.groovy generate \
  -i /local/docs/swagger/openapi.yml \
  -g CollectionCodegen \
  -t /local/generate \
  -o /local


FROM golang:1.14.2-alpine AS build

RUN apk add --update --no-cache git upx tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /Tokyo && \
  apk del tzdata && \
  rm -rf /var/cache/apk/* && \
  mkdir -p /usr/share/zoneinfo/Asia && \
  mv /Tokyo /usr/share/zoneinfo/Asia

WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY go.* ./
RUN go mod download

COPY --from=generate /home/groovy/scripts/ ./
ENTRYPOINT go test ./router