module github.com/halooid/gateway

go 1.22.0

require (
	github.com/halooid/backend/go-shared v0.0.0
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.1
)

replace github.com/halooid/backend/go-shared => ../backend/go-shared
