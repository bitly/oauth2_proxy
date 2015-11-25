FROM golang:1.5.2
MAINTAINER colin.hom@coreos.com

RUN go get github.com/tools/godep

ADD . $GOPATH/src/github.com/bitly/oauth2_proxy

WORKDIR $GOPATH/src/github.com/bitly/oauth2_proxy

RUN godep go install github.com/bitly/oauth2_proxy

ENTRYPOINT ["oauth2_proxy"]
