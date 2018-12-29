FROM golang:alpine as builder
MAINTAINER Alexandre FARIA, alexandre.faria@decathlon.com

# Git is required for fetching the dependencies. (go get ...)
RUN apk update && apk add --no-cache git

COPY . $GOPATH/src/github.com/bitly/oauth2_proxy
WORKDIR $GOPATH/src/github.com/bitly/oauth2_proxy

RUN go get -d -v
ENV CGO_ENABLED=0
RUN go build -o /go/bin/oauth2_proxy

# Create appuser.
RUN adduser -D -g '' decathlon

COPY ./docker/docker-entrypoint.sh ./temp-entrypoint.sh
#Prevent windows developer to inject bad end of lines into the shell script
RUN cat ./temp-entrypoint.sh | tr -d '\r' >  docker-entrypoint.sh
RUN chmod +x docker-entrypoint.sh



FROM alpine

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/oauth2_proxy /usr/local/bin/oauth2_proxy
COPY --from=builder /go/src/github.com/bitly/oauth2_proxy/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# non root user
USER decathlon

EXPOSE 4180
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
#CMD ["--upstream=http://0.0.0.0:8000/"]
