FROM docker.io/schasse/oauth2_proxy:latest
MAINTAINER kraven@kraven.org
COPY dist/oauth2_proxy /go/bin
COPY contrib/oauth2_proxy.cfg.example /etc/oauth2_proxy.cfg
RUN chmod +x /go/bin/oauth2_proxy
CMD["-config", "/etc/oauth2_proxy.cfg"]
ENTRYPOINT ["oauth2_proxy"]
