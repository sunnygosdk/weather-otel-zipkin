module github.com/sunnygosdk/weather-otel-zipkin/serviceb

go 1.25.1

require (
	github.com/gorilla/mux v1.8.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.63.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/zipkin v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)
