FROM golang:1.5
MAINTAINER colin.hom@coreos.com

# Install gpm tool
RUN wget -qO- https://raw.githubusercontent.com/pote/gpm/v1.3.2/bin/gpm > /usr/bin/gpm
RUN chmod +x /usr/bin/gpm

# Preserve cached gpm install unless ./Godeps file is modified
ADD ./Godeps /usr/share/oauth2_proxy/Godeps
WORKDIR /usr/share/oauth2_proxy
RUN gpm install

# add oauth2_proxy to GOPATH
ADD ./ $GOPATH/src/github.com/bitly/oauth2_proxy

# install oauth2_proxy
RUN go install github.com/bitly/oauth2_proxy
ENTRYPOINT [ "oauth2_proxy" ]

# expose default http/https listen ports
EXPOSE 4180
EXPOSE 443
