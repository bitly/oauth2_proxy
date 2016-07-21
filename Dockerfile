FROM golang:1.6-alpine

# install su-exec (a lightweight gosu)
RUN apk add --no-cache su-exec libcap

# install all build dependencies
RUN apk add --no-cache --virtual build-deps jq curl tar git

# create oauth2_proxy user and group
RUN addgroup -S oauth2_proxy && \
    adduser -D -S -s /sbin/nologin -G oauth2_proxy oauth2_proxy


# download the latest oauth2_proxy release
RUN mkdir -p /go/src/app && \
    curl -sSL https://api.github.com/repos/bitly/oauth2_proxy/releases/latest | \
    jq -r .tarball_url | \
    xargs -n 1 curl -sSL | \
    tar -xzf - --strip 1 -C /go/src/app

# get all the dependencies
RUN go get -d -v github.com/bitly/oauth2_proxy && \
    go install -v github.com/bitly/oauth2_proxy

# cleanup
RUN apk del build-deps && \
    rm -rf /var/cache/apk/*

VOLUME /config

# a non root user can't bind to localhost on network='host' mode
# unless app doesn't have the net_bind_service cap
# https://github.com/docker/docker/issues/8460
RUN setcap 'cap_net_bind_service=+ep' /go/bin/oauth2_proxy

EXPOSE 4180

ENTRYPOINT ["su-exec", "oauth2_proxy", "oauth2_proxy"]
CMD ["-config", "/config/oauth2_proxy.cfg"]