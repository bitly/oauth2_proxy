FROM golang:1.11-stretch

WORKDIR /go

RUN go get -d -v github.com/bitly/oauth2_proxy
RUN go install -v github.com/bitly/oauth2_proxy

CMD ["oauth2_proxy"]

# Multi-stage build setup 

# Stage 1 (to create a "build" image, ~850MB)
FROM golang:1.11 AS builder
RUN go version

RUN go get -v github.com/bitly/oauth2_proxy

WORKDIR /go/src/github.com/bitly/oauth2_proxy/
COPY . /go/src/github.com/bitly/oauth2_proxy/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o oauth2_proxy .

# Stage 2

# FROM alpine:3.7
FROM docker.bulogics.com/tools:0.3.0
# RUN apk --no-cache add ca-certificates

WORKDIR /go/bin
COPY --from=builder /go/src/github.com/bitly/oauth2_proxy/oauth2_proxy .

ENTRYPOINT ["./oauth2_proxy"]