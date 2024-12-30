# Makefile to generate Go code from .proto files

# Directory where your .proto files are located
PROTO_DIR=api/v1

# Output directory for the generated Go files
OUT_DIR=api/v1

# List of .proto files
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)

.PHONY: proto clean

# Default target to generate the Go code
proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go_opt=paths=source_relative \
		$(PROTO_FILES)

# Clean up generated files
clean:
	rm -f $(OUT_DIR)/*.pb.go

# compile:
# 	protoc \
# 		--proto_path=$(PROTO_DIR) \
# 		--go_out=$(OUT_DIR) \
# 		--go_opt=paths=source_relative \
# 		--go-grpc_opt=paths=source_relative \
# 		$(PROTO_FILES)
