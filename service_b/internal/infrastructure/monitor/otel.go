package monitor

import (
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracer "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.13.0"
	"go.opentelemetry.io/otel/trace"
)

type OpenTelemetry struct {
	ServiceName      string
	ServiceVersion   string
	ExporterEndpoint string
}

func NewOpenTelemetry() *OpenTelemetry {
	return &OpenTelemetry{}
}

func (o *OpenTelemetry) GetTracer() trace.Tracer {
	var logger = log.New(os.Stderr, "zipkin", log.Ldate|log.Ltime|log.Llongfile)
	exporter, err := zipkin.New(
		o.ExporterEndpoint,
		zipkin.WithLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	batcher := tracer.NewBatchSpanProcessor(exporter)
	tp := tracer.NewTracerProvider(
		tracer.WithSpanProcessor(batcher),
		tracer.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("serviceb"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracer := otel.Tracer("io.opentelemetry.traces.goapp")
	return tracer
}
