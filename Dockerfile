FROM golang:1.6-alpine

ENV http_proxy http://agproxy.agint:8081
ENV https_proxy http://agproxy.agint:8081
ENV HTTP_PROXY http://agproxy.agint:8081
ENV HTTPS_PROXY http://agproxy.agint:8081

# install su-exec (a lightweight gosu)
RUN apk add --no-cache su-exec libcap bash openssl

# install all build dependencies
RUN apk add --no-cache --virtual build-deps jq curl tar git

# create oauth2_proxy user and group
RUN addgroup -S oauth2_proxy && \
    adduser -D -S -s /sbin/nologin -G oauth2_proxy oauth2_proxy

RUN go get -v github.com/brunnels/oauth2_proxy && \
    go install -v github.com/brunnels/oauth2_proxy

# cleanup
RUN apk del build-deps && \
    rm -rf /var/cache/apk/*

RUN rm -rf /usr/local/share/ca-certificates && \
    mkdir /usr/local/share/ca-certificates && \
    export DOMAIN_NAME=gitlab.tools.aig.net && \
    openssl s_client -connect $DOMAIN_NAME:443 -showcerts </dev/null 2>/dev/null | openssl x509 -outform PEM > /usr/local/share/ca-certificates/$DOMAIN_NAME.crt && \
    update-ca-certificates

VOLUME /config

# a non root user can't bind to localhost on network='host' mode
# unless app doesn't have the net_bind_service cap
# https://github.com/docker/docker/issues/8460
RUN setcap 'cap_net_bind_service=+ep' /go/bin/oauth2_proxy

EXPOSE 4180

ENTRYPOINT ["su-exec", "oauth2_proxy", "oauth2_proxy"]
