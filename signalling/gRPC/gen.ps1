$env:PATH += ";${HOME}/go/bin"
.\protoc.exe --go_out=. ./protobuf.proto 