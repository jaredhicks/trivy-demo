package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mitre-tdp/wind-demo-go/internal/api"
)

func newExporter(ctx context.Context) (*stdouttrace.Exporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func newTraceProvider(exp *stdouttrace.Exporter) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"id":1,"content":"Hello, World!"}`)
}

func main() {
	ctx := context.Background()

	exp, err := newExporter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	// Create a new tracer provider with a batch span processor and the exporter.
	tp := newTraceProvider(exp)

	// Handle this error in a sensible manner where possible
	defer func() { _ = tp.Shutdown(ctx) }()

	// Set the Tracer Provider and the W3C Trace Context propagator as globals
	otel.SetTracerProvider(tp)

	// Register the trace context and baggage propagators so data is propagated across services/processes.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// Initialize HTTP handler instrumentation
	// (Wrap the HTTP handler funcs with OTel HTTP instrumentation)
	router := mux.NewRouter()

	router.Handle("/hello", otelhttp.NewHandler(http.HandlerFunc(helloHandler), "root"))

	//
	//router.HandleFunc("/artifacts", api.RealArtifactsHandler).
	//	Methods("POST").
	//	Queries("build", "{buildId}")
	//
	//router.HandleFunc("/artifacts", api.DeleteHandler).
	//	Methods("DELETE")
	//
	router.HandleFunc("/health", api.HealthCheckHandler)

	address := ":8080"

	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server", address)

	log.Fatal(srv.ListenAndServe())
}
