FROM	golang

WORKDIR	/go/src/github.com/bitly/oauth2_proxy
ADD	https://github.com/twhtanghk/oauth2_proxy/archive/master.tar.gz /tmp
RUN	tar --strip-components=1 -xzf /tmp/master.tar.gz && \
	rm /tmp/master.tar.gz && \
        go get && \
        go build -o oauth2_proxy
EXPOSE	4180

ENTRYPOINT ./entrypoint.sh
