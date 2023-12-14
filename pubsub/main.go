package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		slog.Debug(fmt.Sprintf("Defaulting to port %s", port))
	}

	echoServer := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if buf, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, request.Body)); err != nil {

			} else {
				slog.DebugContext(request.Context(), string(buf), "headers", request.Header)

				_, _ = fmt.Fprintln(writer, "Header", request)
				_, _ = fmt.Fprintln(writer, "Body", string(buf))
			}
		}),
	}

	go func() {
		slog.Debug(fmt.Sprintf("Listening on port %s", port))
		if err := echoServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	slog.Debug("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := echoServer.Shutdown(ctx); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
