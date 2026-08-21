# syntax = docker/dockerfile:1

FROM golang:1.27rc3-alpine@sha256:c5aca77a4d16cb6688dbf3ccade67eff6f05ee208bc854d060e6947f5c27e23c AS build

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add tzdata git && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata

WORKDIR /go/src/github.com/traPtitech/trap-collection-server

COPY --from=arigaio/atlas:1.2.0@sha256:69fef5b506378439771fc70e052471780336552dbf35da397f5a7a729da8f7e3 --chown=root:root atlas /usr/local/bin/atlas

COPY go.mod go.sum ./
RUN go mod download

ENV GOFLAGS -buildvcs=false

ENTRYPOINT ["go", "tool", "air"]
CMD ["-c", ".air.toml"]
