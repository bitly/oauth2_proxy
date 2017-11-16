#!/bin/bash
set -e

echo "gofmt"
diff -u <(echo -n) <(gofmt -d $(find . -type f -name '*.go' -not -path "./vendor/*"))
echo "go vet"
go vet $(go list ./... | grep -v /vendor/)
echo "go test"
go test -timeout 60s $(go list ./... | grep -v /vendor/)
echo "go test -race"
GOMAXPROCS=4 go test -timeout 60s -race $(go list ./... | grep -v /vendor/)
