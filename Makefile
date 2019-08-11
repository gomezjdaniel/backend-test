GO_FILES = $(shell find ./ -maxdepth 1 -type f -name '*.go' -and -not -name "*_test.go")

GO_TEST_FILES = $(shell find ./ -maxdepth 1 -type f -name '*_test.go')

dev-deps:
	sudo apt-get install python-pip python-dev libpq-dev libevent-dev
	sudo pip install pgcli

postgres:
	pgcli -d 'postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable'

go.lint:
	revive -formatter friendly $(GO_FILES)

go.test:
	go test -v $(GO_TEST_FILES) $(GO_FILES)

run: build
	./bin/backend-test --log-level=0 --database-url="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" --redis-url="redis://:@localhost:6379/0"

build:
	GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -o ./bin/backend-test

docker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/backend-test-linux-amd64
	docker build -t "gomezjdaniel/backend-test" .
