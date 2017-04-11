FROM golang:1.6-alpine
MAINTAINER brunnels <kraven@kraven.org>

RUN export http_proxy=http://agproxy.agint:8081 && \
    export https_proxy=http://agproxy.agint:8081 && \
    apk add --no-cache --virtual build-deps jq git && \
    go get -v github.com/brunnels/oauth2_proxy && \
    go install -v github.com/brunnels/oauth2_proxy && \
    chmod +x /go/bin/oauth2_proxy && \
    touch /etc/oauth2_proxy.cfg

# cleanup
RUN apk del build-deps && \
    rm -rf /var/cache/apk/* && \
    rm -rf /go/src

CMD ["-config", "/etc/oauth2_proxy.cfg"]
ENTRYPOINT ["oauth2_proxy"]
