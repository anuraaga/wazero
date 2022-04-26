module github.com/tetratelabs/wazero-otel

// This should be the minimum supported Go version per https://go.dev/doc/devel/release (1 version behind latest)
go 1.17

require (
	github.com/tetratelabs/wazero v0.0.0
	go.opentelemetry.io/collector/pdata v0.49.1-0.20220422001137-87ab5de64ce4
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.45.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/tetratelabs/wazero => ../../
