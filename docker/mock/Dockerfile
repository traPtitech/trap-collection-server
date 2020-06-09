FROM groovy:3.0.2 AS generate
WORKDIR /home/groovy/scripts
COPY --chown=groovy:groovy . /local
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

RUN apk --update add tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /Tokyo && \
  apk del tzdata && \
  rm -rf /var/cache/apk/* && \
  mkdir -p /usr/share/zoneinfo/Asia && \
  mv /Tokyo /usr/share/zoneinfo/Asia

COPY --from=build /go/src/github.com/traPtitech/trap-collection-server/main ./
COPY ./upload ./upload

ENTRYPOINT ./main