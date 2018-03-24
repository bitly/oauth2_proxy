FROM golang:alpine

RUN apk add --no-cache git

# Checkout CivicActions' latest google-auth-proxy code from Github
RUN go get github.com/CivicActions/oauth2_proxy

# Expose the ports we need and setup the ENTRYPOINT w/ the default argument
# to be pass in.
EXPOSE 80
ENTRYPOINT [ "./bin/oauth2_proxy" ]
CMD [ "--upstream=http://upstream:80/", "--http-address=0.0.0.0:80" ]
