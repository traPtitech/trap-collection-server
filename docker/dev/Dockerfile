# syntax = docker/dockerfile:1

FROM golang:1.27-alpine@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 AS build

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add tzdata git && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY --from=arigaio/atlas:1.2.0@sha256:69fef5b506378439771fc70e052471780336552dbf35da397f5a7a729da8f7e3 --chown=root:root atlas /usr/local/bin/atlas

COPY go.mod go.sum ./
RUN go mod download

ENV GOFLAGS -buildvcs=false

COPY --from=cosmtrek/air:v1.63.0 /go/bin/air /usr/local/bin/air

ENTRYPOINT ["air"]
CMD ["-c", ".air.toml"]