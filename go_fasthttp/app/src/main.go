package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ex10se/http-perf-test/go_fasthttp/config"
	"github.com/ex10se/http-perf-test/go_fasthttp/handlers"
	"github.com/ex10se/http-perf-test/go_fasthttp/rabbitmq"
	"github.com/valyala/fasthttp"
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
	// Простой роутер для fasthttp
	router := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/status/status/":
			statusHandler.Handle(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}
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
	// Настраиваем fasthttp сервер
	srv := &fasthttp.Server{
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server started on %s", cfg.SocketPath)
		if err := srv.Serve(listener); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	// Ожидаем сигнал завершения
	<-quit
	log.Println("Server is shutting down...")
	// Закрываем listener — fasthttp.Serve завершится с ошибкой net.ErrClosed
	if err := listener.Close(); err != nil {
		log.Printf("Error closing listener: %v", err)
	}
	// Небольшая пауза, чтобы завершить активные запросы (по желанию)
	time.Sleep(1 * time.Second)
	log.Println("Server stopped")
}
