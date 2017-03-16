FROM golang:1.6

ENV http_proxy http://agproxy.agint:8081
ENV https_proxy http://agproxy.agint:8081

RUN go get -v github.com/brunnels/oauth2_proxy && \
    go install -v github.com/brunnels/oauth2_proxy

ENTRYPOINT ["oauth2_proxy"]