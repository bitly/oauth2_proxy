FROM golang:1.9.4-alpine3.7

ENV BIN=oauth2_proxy

COPY build/*-linux-amd64 /go/bin/$BIN

CMD /go/bin/$BIN
