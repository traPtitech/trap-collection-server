# syntax = docker/dockerfile:1

FROM arigaio/atlas:0.37.0 AS atlas

FROM golang:1.25.1-alpine AS build

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add tzdata git && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY --from=atlas --chown=root:root atlas /usr/local/bin/atlas

COPY go.mod go.sum ./
RUN go mod download

ENV GOFLAGS -buildvcs=false

ENTRYPOINT ["go", "tool", "air"]
CMD ["-c", ".air.toml"]
