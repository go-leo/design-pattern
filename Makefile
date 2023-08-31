.PHONY: go_gen
go_gen:
	@echo "--- go generate start ---"
	@go generate ./...
	@echo "--- go generate end ---"

protoc_cqrs_gen:
	@echo "--- protoc generate start ---"
	@protoc \
		--proto_path=. \
		--go_out=. \
		--go_opt=module=github.com/go-leo/design-pattern \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/go-leo/design-pattern \
		--go-cqrs_out=. \
		--go-cqrs_opt=module=github.com/go-leo/design-pattern \
		cqrs/cmd/example/api/pb/*.proto
	@echo "--- protoc generate end ---"