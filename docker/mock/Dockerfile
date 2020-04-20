FROM groovy:3.0.2 AS generate
WORKDIR /home/groovy/scripts
COPY . .
RUN groovy ./generate/generator.groovy generate \
      -i ./docs/swagger/openapi.yml \
      -g CollectionCodegen \
      -t ./generate \
      -o .

FROM golang:1.14.2-alpine AS build
RUN apk add --update --no-cache git
RUN go get github.com/golang/mock/mockgen
WORKDIR /go/src/github.com/traPtitech/trap-collection-server
COPY --from=generate /home/groovy/scripts/ ./
RUN sh mockgen.sh
RUN go mod download
RUN go build -o main

FROM alpine:3.11.5 AS runtime
ENV DOCKERIZE_VERSION v0.6.1
RUN apk add --update --no-cache tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz && \
  rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/main ./
ENTRYPOINT dockerize -wait tcp://mariadb:3306 ./main