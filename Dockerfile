FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /go/src/oauth2_proxy
COPY . .

RUN go get -v && go install -v

# Expose the ports we need and setup the ENTRYPOINT w/ the default argument
# to be pass in.
EXPOSE 80
ENTRYPOINT [ "/go/bin/oauth2_proxy" ]
CMD [ "--upstream=http://upstream:80/", "--http-address=0.0.0.0:80" ]
