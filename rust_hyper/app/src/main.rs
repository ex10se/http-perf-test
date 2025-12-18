mod config;
mod handlers;
mod models;
mod rabbitmq;

use http_body_util::Full;
use hyper::body::{Bytes, Incoming};
use hyper::server::conn::http1;
use hyper::service::service_fn;
use hyper::{Request, Response};
use hyper_util::rt::TokioIo;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::sync::Arc;
use tokio::net::UnixListener;
use tokio::signal;
use tracing::info;

use crate::config::Config;
use crate::handlers::{status_handler, AppState};
use crate::rabbitmq::RabbitClient;

async fn handle_request(
    state: Arc<AppState>,
    req: Request<Incoming>,
) -> Result<Response<Full<Bytes>>, Box<dyn std::error::Error + Send + Sync>> {
    status_handler(state, req).await
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
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

    // Удаляем старый socket если существует
    let _ = fs::remove_file(&config.socket_path);

    // Создаем Unix socket listener
    let listener = UnixListener::bind(&config.socket_path)?;

    // Устанавливаем права на socket
    fs::set_permissions(&config.socket_path, fs::Permissions::from_mode(0o666))?;

    info!("Server started successfully on {}", config.socket_path);

    // Обрабатываем входящие соединения с graceful shutdown
    tokio::select! {
        _ = async {
            loop {
                let (stream, _) = listener.accept().await.expect("Failed to accept");
                let io = TokioIo::new(stream);
                let state = app_state.clone();

                tokio::spawn(async move {
                    let service = service_fn(move |req| {
                        let state = state.clone();
                        handle_request(state, req)
                    });

                    if let Err(err) = http1::Builder::new().serve_connection(io, service).await {
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
    Ok(())
}
