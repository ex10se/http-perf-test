package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ex10se/http-perf-test/golang_nginx/config"
	"github.com/ex10se/http-perf-test/golang_nginx/handlers"
	"github.com/ex10se/http-perf-test/golang_nginx/rabbitmq"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()
	log.Printf("Starting server on %s", cfg.SocketPath)
	log.Printf("RabbitMQ URL: %s", cfg.RabbitMQURL)
	// Создаем RabbitMQ клиент
	rmqClient := rabbitmq.New(cfg.RabbitMQURL)
	defer func() {
		if err := rmqClient.Close(); err != nil {
			log.Printf("Error closing RabbitMQ client: %v", err)
		}
	}()
	// Декларируем очереди
	log.Println("Declaring RabbitMQ queues...")
	if err := rmqClient.DeclareQueues(); err != nil {
		log.Fatalf("Failed to declare queues: %v", err)
	}
	// Создаем HTTP хэндлер
	statusHandler := handlers.NewStatusHandler(rmqClient)
	// Настраиваем роутинг
	mux := http.NewServeMux()
	mux.Handle("/status/status/", statusHandler)
	// Создаем Unix socket
	if err := os.RemoveAll(cfg.SocketPath); err != nil {
		log.Fatalf("Failed to remove old socket: %v", err)
	}
	listener, err := net.Listen("unix", cfg.SocketPath)
	if err != nil {
		log.Fatalf("Failed to create socket: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}()
	// Устанавливаем права на socket
	if err := os.Chmod(cfg.SocketPath, 0666); err != nil {
		log.Fatalf("Failed to chmod socket: %v", err)
	}
	// Настраиваем HTTP сервер
	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Канал для graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server started on %s", cfg.SocketPath)
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	// Ожидаем сигнал завершения
	<-quit
	log.Println("Server is shutting down...")
	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	done <- true
	log.Println("Server stopped gracefully")
}
