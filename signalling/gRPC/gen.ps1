$env:PATH += ";${HOME}/go/bin"
.\protoc.exe --go_out=. --go-grpc_out=. ./protobuf.proto 