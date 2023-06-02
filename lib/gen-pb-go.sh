protoc --proto_path=./proto \
  --go_out=./helloworld --go_opt=paths=source_relative \
  --go-grpc_out=./helloworld --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=./helloworld  --grpc-gateway_opt=paths=source_relative \
  ./proto/helloworld/helloworld.proto
