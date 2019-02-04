#!/bin/bash

while [[ $# -gt 0 ]]
do
    key="$1"
    case $key in
        -v|--verbose)
        VERBOSE="-v"
        echo "Verbose flag is set. Expect noisy output"
        ;;
    esac
done

EXIT_CODE=0
echo "gofmt:"
diff -u <(echo -n) <(gofmt -d $(find . -type f -name '*.go' -not -path "./vendor/*")) || EXIT_CODE=1
for pkg in $(go list ./... | grep -v '/vendor/' ); do
    echo ""
    echo "testing $pkg:"
    echo "go vet $pkg"
    go vet "$pkg" || EXIT_CODE=1

    echo "go test ${VERBOSE} $pkg"
    go test ${VERBOSE} -timeout 90s "$pkg" || EXIT_CODE=1

    echo "go test ${VERBOSE} -race $pkg"
    GOMAXPROCS=4 go test ${VERBOSE} -timeout 90s0s -race "$pkg" || EXIT_CODE=1
done
exit $EXIT_CODE