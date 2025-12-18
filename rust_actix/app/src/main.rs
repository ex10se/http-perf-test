mod config;
mod handlers;
mod models;
mod rabbitmq;

use actix_web::{middleware, web, App, HttpServer};
use log::info;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::sync::Arc;
use tokio::signal;

use crate::config::Config;
use crate::handlers::{status_handler, AppState};
use crate::rabbitmq::RabbitClient;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::new().default_filter_or("info"));

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

    info!("Server starting on {}", config.socket_path);

    // Создаем HTTP сервер
    let server = HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(app_state.clone()))
            // Logger middleware убран для максимальной производительности
            .route("/status/status/", web::post().to(status_handler))
    })
    .workers(10)
    .bind_uds(&config.socket_path)?;

    // Устанавливаем права на socket
    fs::set_permissions(&config.socket_path, fs::Permissions::from_mode(0o666))?;

    info!("Server started successfully");

    // Запускаем сервер
    let server_handle = server.run();
    
    // Graceful shutdown при получении сигнала
    tokio::select! {
        _ = server_handle => {},
        _ = signal::ctrl_c() => {
            info!("Shutting down gracefully...");
            if let Err(e) = rabbit_client.close().await {
                log::error!("Error closing RabbitMQ connection: {}", e);
            }
        }
    }

    info!("Server stopped");
    Ok(())
}
