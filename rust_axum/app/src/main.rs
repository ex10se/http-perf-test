mod config;
mod handlers;
mod models;
mod rabbitmq;

use axum::{routing::post, Router};
use hyper::body::Incoming;
use hyper::Request;
use hyper_util::rt::TokioIo;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::sync::Arc;
use tokio::signal;
use tower::Service;
use tower_http::trace::TraceLayer;
use tracing::info;

use crate::config::Config;
use crate::handlers::{status_handler, AppState};
use crate::rabbitmq::RabbitClient;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info")),
        )
        .init();

    let config = Config::from_env();
    info!("Starting server on {}", config.socket_path);
    info!("RabbitMQ URL: {}", config.rabbitmq_url);

    // Создаем RabbitMQ клиент
    let rabbit_client = RabbitClient::new(config.rabbitmq_url.clone());

    // Инициализируем соединение
    info!("Connecting to RabbitMQ...");
    rabbit_client
        .init()
        .await
        .expect("Failed to connect to RabbitMQ");

    // Декларируем очереди
    info!("Declaring RabbitMQ queues...");
    rabbit_client
        .declare_queues()
        .await
        .expect("Failed to declare queues");

    // Создаем shared state
    let app_state = Arc::new(AppState {
        rabbit_client: rabbit_client.clone(),
    });

    // Создаем роутер
    let app = Router::new()
        .route("/status/status/", post(status_handler))
        // TraceLayer убран для максимальной производительности
        .with_state(app_state);

    // Удаляем старый socket если существует
    let _ = fs::remove_file(&config.socket_path);

    info!("Server starting on {}", config.socket_path);

    // Создаем Unix socket listener
    let listener = tokio::net::UnixListener::bind(&config.socket_path)
        .expect("Failed to bind unix socket");

    // Устанавливаем права на socket
    fs::set_permissions(&config.socket_path, fs::Permissions::from_mode(0o666))
        .expect("Failed to set socket permissions");

    info!("Server started successfully");

    // Обрабатываем входящие соединения с graceful shutdown
    tokio::select! {
        _ = async {
            loop {
                let (stream, _) = listener.accept().await.expect("Failed to accept connection");
                let tower_service = app.clone();

                tokio::spawn(async move {
                    let socket = TokioIo::new(stream);
                    let hyper_service = hyper::service::service_fn(move |request: Request<Incoming>| {
                        tower_service.clone().call(request)
                    });

                    if let Err(err) = hyper_util::server::conn::auto::Builder::new(hyper_util::rt::TokioExecutor::new())
                        .serve_connection(socket, hyper_service)
                        .await
                    {
                        eprintln!("Error serving connection: {:?}", err);
                    }
                });
            }
        } => {},
        _ = signal::ctrl_c() => {
            info!("Shutting down gracefully...");
            if let Err(e) = rabbit_client.close().await {
                tracing::error!("Error closing RabbitMQ connection: {}", e);
            }
        }
    }

    info!("Server stopped");
}
