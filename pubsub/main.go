package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
			var buf strings.Builder
			_, _ = io.Copy(&buf, request.Body)
			slog.DebugContext(request.Context(), buf.String(), "Ce-Id", request.Header.Get("Ce-Id"))

			_, _ = fmt.Fprintln(writer, "Ce-Id: %s", request.Header.Get("Ce-Id"))
			_, _ = fmt.Fprintln(writer, buf.String())
		}),
	}

	go func() {
		slog.Debug(fmt.Sprintf("Listening on port %s", port))
		if err := echoServer.ListenAndServe(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := echoServer.Shutdown(ctx); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
