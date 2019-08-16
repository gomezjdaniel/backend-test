FROM alpine:3.10
ADD ./bin/backend-test-linux-amd64 /go/bin/backend-test
ENTRYPOINT ["/go/bin/backend-test"]
