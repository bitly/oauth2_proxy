# build stage
FROM golang:1.9-stretch AS build-env
WORKDIR /go/src/github.com/bitly/oauth2_proxy
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

# final stage
FROM debian:stretch-slim
WORKDIR /app
COPY --from=build-env /go/bin/oauth2_proxy /app/

# Install CA certificates
RUN apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates

EXPOSE 8080 4180
ENTRYPOINT [ "./oauth2_proxy" ]
CMD [ "--upstream=http://127.0.0.1:8080/", "--http-address=0.0.0.0:4180" ]
