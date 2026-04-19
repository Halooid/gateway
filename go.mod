module github.com/halooid/gateway

go 1.24.0

require github.com/halooid/backend/go-shared v0.0.0

require (
	github.com/MicahParks/keyfunc/v2 v2.1.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/halooid/backend/go-shared => ../backend/go-shared
