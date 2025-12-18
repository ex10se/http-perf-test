use flate2::write::GzEncoder;
use flate2::Compression;
use lapin::{
    options::*,
    types::FieldTable,
    BasicProperties, Channel, Connection, ConnectionProperties,
};
use std::io::Write;
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing::{error, info};

use super::queues::{EXCHANGE_NAME, QUEUE_AXUM, QUEUE_SYSTEM_AXUM};

#[derive(Clone)]
pub struct RabbitClient {
    channel: Arc<Mutex<Option<Channel>>>,
    url: String,
}

impl RabbitClient {
    pub fn new(url: String) -> Self {
        Self {
            channel: Arc::new(Mutex::new(None)),
            url,
        }
    }

    pub async fn init(&self) -> Result<(), Box<dyn std::error::Error>> {
        info!("Connecting to RabbitMQ: {}", self.url);
        
        let conn = Connection::connect(&self.url, ConnectionProperties::default()).await?;
        let channel = conn.create_channel().await?;

        info!("Connected to RabbitMQ");

        // Декларируем exchange
        channel
            .exchange_declare(
                EXCHANGE_NAME,
                lapin::ExchangeKind::Direct,
                ExchangeDeclareOptions {
                    durable: true,
                    ..Default::default()
                },
                FieldTable::default(),
            )
            .await?;

        info!("Exchange '{}' declared", EXCHANGE_NAME);

        *self.channel.lock().await = Some(channel);
        Ok(())
    }

    pub async fn declare_queues(&self) -> Result<(), Box<dyn std::error::Error>> {
        let channel_lock = self.channel.lock().await;
        let channel = channel_lock
            .as_ref()
            .ok_or("Channel not initialized")?;

        // Декларируем очередь для обычных событий
        channel
            .queue_declare(
                QUEUE_AXUM,
                QueueDeclareOptions {
                    durable: true,
                    ..Default::default()
                },
                FieldTable::default(),
            )
            .await?;

        channel
            .queue_bind(
                QUEUE_AXUM,
                EXCHANGE_NAME,
                QUEUE_AXUM,
                QueueBindOptions::default(),
                FieldTable::default(),
            )
            .await?;

        info!("Queue '{}' declared and bound", QUEUE_AXUM);

        // Декларируем очередь для системных событий
        channel
            .queue_declare(
                QUEUE_SYSTEM_AXUM,
                QueueDeclareOptions {
                    durable: true,
                    ..Default::default()
                },
                FieldTable::default(),
            )
            .await?;

        channel
            .queue_bind(
                QUEUE_SYSTEM_AXUM,
                EXCHANGE_NAME,
                QUEUE_SYSTEM_AXUM,
                QueueBindOptions::default(),
                FieldTable::default(),
            )
            .await?;

        info!("Queue '{}' declared and bound", QUEUE_SYSTEM_AXUM);

        Ok(())
    }

    pub async fn publish(&self, routing_key: &str, body: &[u8]) -> Result<(), Box<dyn std::error::Error>> {
        // Компрессия ВНЕ mutex (как в Go)
        let compressed = compress_message(body)?;

        // Минимальное время удержания mutex
        let channel_lock = self.channel.lock().await;
        let channel = channel_lock
            .as_ref()
            .ok_or("Channel not initialized")?;

        // Отправляем и НЕ ждем подтверждения (быстрее)
        channel
            .basic_publish(
                EXCHANGE_NAME,
                routing_key,
                BasicPublishOptions::default(),
                &compressed,
                BasicProperties::default()
                    .with_content_type("application/json".into())
                    .with_content_encoding("gzip".into())
                    .with_delivery_mode(2), // Persistent
            )
            .await?;
        // НЕ вызываем второй .await - не ждем подтверждения

        Ok(())
    }

    pub async fn close(&self) -> Result<(), Box<dyn std::error::Error>> {
        let mut channel_lock = self.channel.lock().await;
        if let Some(channel) = channel_lock.take() {
            channel.close(200, "Normal shutdown").await?;
            info!("RabbitMQ channel closed");
        }
        Ok(())
    }
}

/// Сжимает сообщение с помощью gzip (как в Go)
fn compress_message(data: &[u8]) -> Result<Vec<u8>, std::io::Error> {
    let mut encoder = GzEncoder::new(Vec::new(), Compression::default());
    encoder.write_all(data)?;
    encoder.finish()
}
