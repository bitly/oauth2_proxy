FROM golang:latest as builder

# Install dep
RUN go get -u github.com/golang/dep/cmd/dep

# Install upx, a Linux binary compression util
RUN apt-get update && apt-get install -y upx

WORKDIR /go/src
COPY . github.com/webflow/oauth2_proxy
WORKDIR /go/src/github.com/webflow/oauth2_proxy

# Load pinned dependencies into vendor/
RUN dep ensure -v

# Build and strip our binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=`git log --pretty=format:'%h' -n 1`" -a -installsuffix cgo -o oauth2_proxy .

# Compress the binary with upx
RUN upx oauth2_proxy

FROM ubuntu

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y dist-upgrade

# Copy the binary over from the builder image
COPY --from=builder /go/src/github.com/webflow/oauth2_proxy/oauth2_proxy /

# Run our entrypoint script when the container is executed
CMD ["/oauth2_proxy"]
