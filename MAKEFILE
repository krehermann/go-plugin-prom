

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	api/v1/controller/controller.proto

.PHONY: plugin
plugin:
	go build -o ./bin/greeter ./plugin/greeter.go

.PHONY: cli
cli:
	go build -o ./bin/cli ./cmd/cli/root.go 

.PHONY: all
all: cli plugin