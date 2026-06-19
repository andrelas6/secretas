package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrelas6/secretas/internal/env"
	"github.com/andrelas6/secretas/internal/k8s/health"
	"github.com/andrelas6/secretas/internal/observability"
	"github.com/andrelas6/secretas/internal/secret/controller"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

var (
	logger = otelslog.NewLogger("main")
)

func run() (err error) {
	// Interrupt - Handle SIGINT (CTRL+C) gracefully on all OS platforms
	// syscall.SIGTERM - Linux specific
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	// env vars
	env, err := env.NewEnv(".env")

	if err != nil {
		return err
	}

	otelShutdown, err := observability.SetupOTelSDK(ctx)
	if err != nil {
		return err
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = errors.Join(err, otelShutdown(shutdownCtx))
	}()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cmp.Or(env.GetEnv("PORT"), "3001")),
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHttpHandler(),
	}

	srvErr := make(chan error, 1)
	go func() {
		logger.Info("Starting server on localhost:3001")
		srvErr <- srv.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		// error when starting http server
		fmt.Println("Error when starting http server")
		return err
	case <-ctx.Done():
		// wait for first ctrl + c
		// stop receiving signal notifications asap
		fmt.Println("stopping server after signal")
		stop()
	}

	// when shutdown is called, listenandserve immediately returns ErrServerClosed
	fmt.Println("server shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)

	return err
}

func newHttpHandler() http.Handler {
	mux := http.NewServeMux()
	secretHandler := controller.SecretHandler{}
	healthzHandler := health.HealthzHandler{}

	mux.Handle("/secret", secretHandler)
	mux.Handle("/healthz", healthzHandler)

	handler := otelhttp.NewHandler(mux, "vaultkit-http")

	return handler
}
