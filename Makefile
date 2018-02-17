VERSION ?= $(shell git describe --always --tags)
BIN = oauth2_proxy
BUILD_CMD = go build -o build/$(BIN)-$(VERSION)-$${GOOS}-$${GOARCH} &
IMAGE_REPO = docker.io

default:
	$(MAKE) bootstrap
	$(MAKE) build

test:
	go vet ./...
	golint -set_exit_status $(shell go list ./... | grep -v vendor)
	go test -covermode=atomic -race -v ./...
bootstrap:
	dep ensure
build:
	go build -o $(BIN)
clean:
	rm -rf build vendor
	rm -f release image bootstrap $(BIN)
release: bootstrap
	@echo "Running build command..."
	bash -c '\
		export GOOS=linux; export GOARCH=amd64; export CGO_ENABLED=0; $(BUILD_CMD) \
		wait \
	'
	touch release

image: release
	@echo "Building the Docker image..."
	docker build -t $(IMAGE_REPO)/$(BIN):$(VERSION) .
	docker tag $(IMAGE_REPO)/$(BIN):$(VERSION) $(IMAGE_REPO)/$(BIN):latest
	touch image

image-push: image
	docker push $(IMAGE_REPO)/$(BIN):$(VERSION)
	docker push $(IMAGE_REPO)/$(BIN):latest

.PHONY: test build clean image-push

