module github.com/leemingeer/grpc-simple

go 1.19

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.2
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.40.0
