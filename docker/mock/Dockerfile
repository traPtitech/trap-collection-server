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

RUN apk add --update --no-cache git upx

WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY go.* ./
RUN go mod download

COPY --from=generate /local/ ./
RUN go build -o main -ldflags "-s -w"
RUN upx ./main


FROM alpine:3.11.6 AS runtime

ENV DOCKERIZE_VERSION v0.6.1
RUN apk --update add tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /Tokyo && \
  apk del tzdata && \
  rm -rf /var/cache/apk/* && \
  mkdir -p /usr/share/zoneinfo/Asia && \
  mv /Tokyo /usr/share/zoneinfo/Asia
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/main ./
COPY ./upload ./upload

ENTRYPOINT dockerize -wait tcp://collection-mariadb:3306 ./main