FROM golang:1.12.7-alpine
RUN apk add --update --no-cache ca-certificates git

WORKDIR /work
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o app
ENTRYPOINT ./app