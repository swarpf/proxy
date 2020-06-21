protoc -I="./proto/api" --go_out=proto-gen/ ./proto/api/proxyapi.proto
protoc -I="./proto/api" --go-grpc_out=proto-gen/ ./proto/api/proxyapi.proto