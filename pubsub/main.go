package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type PubSubMessage struct {
	Message struct {
		Data        []byte            `json:"data,omitempty"`
		MessageId   string            `json:"messageId"`
		PublishTime time.Time         `json:"publishTime"`
		Attributes  map[string]string `json:"attributes,omitempty"`
		OrderingKey string            `json:"orderingKey,omitempty"`
	} `json:"message"`
	Subscription string `json:"subscription"`
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
			var e PubSubMessage
			if err := json.NewDecoder(request.Body).Decode(&e); err != nil {
				http.Error(writer, "Bad HTTP Request", http.StatusBadRequest)
				slog.Error(fmt.Sprintf("Bad HTTP Request: %v", http.StatusBadRequest), "error", err)
				return
			}
			slog.DebugContext(request.Context(), "Event Received",
				"headers", request.Header,
				"event", e,
				"data", string(e.Message.Data),
			)

			_, _ = fmt.Fprintln(writer, "Header", request.Header)
			_, _ = fmt.Fprintln(writer, "event", e)
			_, _ = fmt.Fprintln(writer, "data", string(e.Message.Data))
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
