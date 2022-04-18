module integrations

go 1.17

require (
	github.com/tetratelabs/wazero v0.0.0
	go.opentelemetry.io/otel v1.6.3
	go.opentelemetry.io/otel/exporters/zipkin v1.6.3
	go.opentelemetry.io/otel/sdk v1.6.3
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/openzipkin/zipkin-go v0.4.0 // indirect
	go.opentelemetry.io/otel/trace v1.6.3 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
)

replace github.com/tetratelabs/wazero => ../../
