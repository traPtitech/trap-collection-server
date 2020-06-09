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

ENV DOCKERIZE_VERSION v0.6.1
RUN apk add --update --no-cache git && \
  apk --update add tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata && \
  rm -rf /var/cache/apk/*
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
RUN GO111MODULE=off go get github.com/oxequa/realize && \
  cd /go/src/github.com/oxequa/realize && \
  git fetch && \
  git checkout v2.0.2 && \
  go get github.com/oxequa/realize

WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY go.* ./
RUN go mod download
COPY --from=generate /home/groovy/scripts/ ./
RUN go build -o main

ENTRYPOINT dockerize -wait tcp://mariadb:3306 realize start --name='server' --install --run
