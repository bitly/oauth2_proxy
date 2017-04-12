FROM golang:1.6-alpine
MAINTAINER brunnels <kraven@kraven.org>

RUN apk add --no-cache --virtual build-deps jq git && \
    go get -v github.com/brunnels/oauth2_proxy && \
    go install -v github.com/brunnels/oauth2_proxy && \
    chmod +x /go/bin/oauth2_proxy && \
    touch /etc/oauth2_proxy.cfg

# cleanup
RUN apk del build-deps && \
    rm -rf /var/cache/apk/* && \
    rm -rf /go/src

EXPOSE 4180

ENTRYPOINT ["oauth2_proxy"]
