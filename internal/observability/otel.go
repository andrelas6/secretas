package observability

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// docs: https://opentelemetry.io/docs/languages/go/getting-started/
func SetupOTelSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error
	// later logging and metrics will be here
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// -- setup tracing
	tracerProvider, err := newTraceProvider(ctx, "vaultkit-api")

	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// -- setup metrics
	meterProvider, err := newMeterProvider(ctx, "vaultkit-api")
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// -- setup logs
	loggerProvider, err := newLoggerProvider(ctx, "vaultkit-api")
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

// as other functions need files, extract this to its own package
func newOtelFile(filename string) (*os.File, error) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("observability.newTracerFile: Could not get wd", err)
		return nil, err
	}

	if err := os.MkdirAll(filepath.Join(wd, ".otel"), 0o755); err != nil {
		fmt.Println("observability.newTracerFile: Could not create .otel dir", err)
		return nil, err
	}

	f, err := os.OpenFile(
		filepath.Join(wd, ".otel", filename),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)

	if err != nil {
		fmt.Println("observability.newTracerFile: Could not create file", err)
		return nil, err
	}

	return f, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newMeterProvider(ctx context.Context, serviceName string) (*metric.MeterProvider, error) {
	file, err := newOtelFile("metrics.json")
	if err != nil {
		fmt.Println("observability.newMeterProvider: could not create metrics file", err)
		return nil, err
	}

	metricExporter, err := stdoutmetric.New(stdoutmetric.WithWriter(file))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("create metric resource: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(metricExporter,
				// Default is 1m. Set to 3s for demonstrative purposes.
				metric.WithInterval(3*time.Second),
			),
		),
		metric.WithResource(res),
	)
	return meterProvider, nil
}

func newTraceProvider(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
	// traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())

	file, err := newOtelFile("traces.json")
	if err != nil {
		return nil, err
	}

	traceExporter, err := stdouttrace.New(stdouttrace.WithWriter(file))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("create trace resource: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	return tracerProvider, nil
}

func newLoggerProvider(ctx context.Context, serviceName string) (*log.LoggerProvider, error) {
	file, err := newOtelFile("logs.json")
	if err != nil {
		fmt.Println("observability.newLoggerProvider: could not create logs.json file", err)
		return nil, err
	}

	logExporterFile, err := stdoutlog.New(
		stdoutlog.WithWriter(file),
	)

	if err != nil {
		return nil, err
	}

	logExporterStdOut, err := stdoutlog.New(
		stdoutlog.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("create logger resource: %w", err)
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporterFile)),
		log.WithProcessor(log.NewBatchProcessor(logExporterStdOut)),
		log.WithResource(res),
	)
	return loggerProvider, nil
}
