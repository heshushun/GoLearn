GOPATH:=$(shell go env GOPATH)
MODIFY=Mproto/imports/api.proto=github.com/micro/go-micro/v2/api/proto

.PHONY: proto
proto:

	protoc --proto_path=. --micro_out=${MODIFY}:. --go_out=${MODIFY}:. proto/task/task.proto
	protoc-go-inject-tag -input=proto/task/task.pb.go

.PHONY: build
build: proto

	go build -o task-srv main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build . -t task-srv:latest
