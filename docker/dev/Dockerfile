# syntax = docker/dockerfile:1

FROM arigaio/atlas:0.33.0 AS atlas

FROM golang:1.24.3-alpine AS build

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add tzdata git && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY --from=atlas --chown=root:root atlas /usr/local/bin/atlas

COPY go.mod go.sum ./
RUN go mod download

ENV GOFLAGS -buildvcs=false

ENTRYPOINT ["go", "run", "github.com/cosmtrek/air"]
CMD ["-c", ".air.toml"]
