FROM golang:1.12.7-alpine
RUN apk add --update --no-cache ca-certificates git

WORKDIR /work
COPY . .
RUN source ./env.sh
RUN go build -o app
ENTRYPOINT ./app