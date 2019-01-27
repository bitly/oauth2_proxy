FROM golang:1.11.0 AS build
LABEL maintainer "Fawad Halim <fawad@fawad.net>"

WORKDIR /build
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /build/oauth2_proxy

FROM scratch
COPY --from=build /build/oauth2_proxy /usr/local/bin/oauth2_proxy
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080 4180
ENTRYPOINT [ "/usr/local/bin/oauth2_proxy" ]
CMD [ "--upstream=http://0.0.0.0:8080/", "--http-address=0.0.0.0:4180" ]