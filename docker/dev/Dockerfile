# syntax = docker/dockerfile:1

FROM golang:1.27.3-alpine@sha256:f2c872bcee949b7be9c490efd02b48a8a2fbf947cdf533f1c58525ec9d8170fd AS build

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