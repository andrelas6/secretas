package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/andrelas6/secretas/internal/observability"
	"github.com/andrelas6/secretas/internal/secret/controller"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type rootHandler struct {
}

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Message: "hello",
		Status:  200,
	})
}

func main() {
	// mux := http.NewServeMux()
	// r := rootHandler{}
	// h := controller.SecretHandler{}

	// mux.Handle("/", r)
	// mux.Handle("/secret", h)

	// fmt.Println("Server running on port 3001")
	// if err := http.ListenAndServe(":3001", mux); err != nil {
	// 	fmt.Println("could not start the server", err)
	// 	os.Exit(1)
	// }
	//
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()

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
		Addr:         ":3001",
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHttpHandler(),
	}

	srvErr := make(chan error, 1)
	go func() {
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

	mux.Handle("/secret", secretHandler)

	handler := otelhttp.NewHandler(mux, "vaultkit-http")

	return handler
}
