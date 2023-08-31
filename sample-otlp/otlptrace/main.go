package main

import (
	"context"
	"git.ddxq.mobi/css-oss-internal/otlptracegrpc"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
	"strconv"
	"strings"
	"time"
)

func newGRPCExporter(ctx context.Context, endpoint string, additionalOpts ...otlptracegrpc.Option) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithReconnectionPeriod(50 * time.Millisecond),
	}

	opts = append(opts, additionalOpts...)
	client := otlptracegrpc.NewClient(opts...)
	return otlptrace.New(ctx, client)
}

func main() {
	ctx := context.Background()
	exp, err := newGRPCExporter(ctx, "10.20.24.88:5317")
	if err != nil {
		log.Fatal(err)
	}

	var attrs = []attribute.KeyValue{
		attribute.String("appname", "vm-kang"),
		attribute.String("env_name", "test"),
		attribute.String("ip", "10.24.153.137"),
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(
			exp,
			// add following two options to ensure flush
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(1),
		),
		sdktrace.WithResource(resource.NewSchemaless(attrs...)),
		sdktrace.WithSpanProcessor(&SpanAttributesProcessor{}),
		sdktrace.WithIDGenerator(New("vm-kang")),
	)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	uid := strings.ReplaceAll(uuid.NewString(), "-", "")
	//tid, _ := trace.TraceIDFromHex(uid)
	tid := "vm-kang^^" + uid + "|" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	println("trace_id:", tid)

	tr := tp.Tracer("CUSTOM")
	testKvs := []attribute.KeyValue{
		attribute.Int("Int", 1),
		attribute.Int64("Int64", int64(3)),
		attribute.Float64("Float64", 2.22),
		attribute.Bool("Bool", true),
		attribute.String("String", "test"),
	}

	uid = strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	//sid, _ := trace.SpanIDFromHex(uid)
	sid := "vm-kang:1|" + uid
	println("span_id:", sid)

	var cfg = trace.SpanContextConfig{
		TraceID: []byte(tid),
		//SpanID:  []byte(sid),
	}

	sc := trace.NewSpanContext(cfg)

	sc1, span := tr.Start(trace.ContextWithRemoteSpanContext(ctx, sc), "f1")
	span.SetAttributes(testKvs...)
	defer span.End()

	_, span1 := tr.Start(sc1, "f2", trace.WithSpanKind(trace.SpanKindProducer), WithAsync())
	span1.End()

	_, span2 := tr.Start(sc1, "f3", trace.WithSpanKind(trace.SpanKindProducer))
	span2.End()

	time.Sleep(time.Second * 5)
}

func WithAsync() trace.SpanStartOption {
	return trace.WithAttributes(attribute.Bool("_ddmc_async", true))
}

type SpanAttributesProcessor struct {
}

func (c *SpanAttributesProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
}

func (c *SpanAttributesProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
}

func (c *SpanAttributesProcessor) Shutdown(ctx context.Context) error {
	return nil
}

func (c *SpanAttributesProcessor) ForceFlush(ctx context.Context) error {
	return nil
}
