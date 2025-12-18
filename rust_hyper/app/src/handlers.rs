use http_body_util::{BodyExt, Full};
use hyper::body::Bytes;
use hyper::{Method, Request, Response, StatusCode};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tracing::error;

use crate::models::StatusEvent;
use crate::rabbitmq::{get_queue_name, RabbitClient};

#[derive(Clone)]
pub struct AppState {
    pub rabbit_client: RabbitClient,
}

#[derive(Serialize)]
struct SuccessResponse {
    status: String,
    processed: usize,
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
}

#[derive(Serialize, Deserialize)]
struct EventError {
    event: String,
    error: String,
}

#[derive(Serialize)]
struct PartialSuccessResponse {
    status: String,
    processed: usize,
    errors: Vec<EventError>,
}

fn json_response<T: Serialize>(status: StatusCode, body: &T) -> Response<Full<Bytes>> {
    let json = serde_json::to_string(body).unwrap_or_else(|_| "{}".to_string());
    Response::builder()
        .status(status)
        .header("Content-Type", "application/json")
        .body(Full::new(Bytes::from(json)))
        .unwrap()
}

pub async fn status_handler(
    state: Arc<AppState>,
    req: Request<hyper::body::Incoming>,
) -> Result<Response<Full<Bytes>>, Box<dyn std::error::Error + Send + Sync>> {
    // Проверяем метод
    if req.method() != Method::POST {
        return Ok(json_response(
            StatusCode::METHOD_NOT_ALLOWED,
            &ErrorResponse {
                error: "Method not allowed".to_string(),
            },
        ));
    }

    // Проверяем путь
    if req.uri().path() != "/status/status/" {
        return Ok(json_response(
            StatusCode::NOT_FOUND,
            &ErrorResponse {
                error: "Not found".to_string(),
            },
        ));
    }

    // Читаем тело запроса
    let body_bytes = req
        .into_body()
        .collect()
        .await?
        .to_bytes();

    if body_bytes.is_empty() {
        return Ok(json_response(
            StatusCode::BAD_REQUEST,
            &ErrorResponse {
                error: "Request body is required".to_string(),
            },
        ));
    }

    // Парсим JSON
    let events: Vec<StatusEvent> = match serde_json::from_slice(&body_bytes) {
        Ok(e) => e,
        Err(err) => {
            error!("Failed to parse JSON: {}", err);
            return Ok(json_response(
                StatusCode::BAD_REQUEST,
                &ErrorResponse {
                    error: "Request body must be a JSON array".to_string(),
                },
            ));
        }
    };

    if events.is_empty() {
        return Ok(json_response(
            StatusCode::BAD_REQUEST,
            &ErrorResponse {
                error: "Request body must contain at least one event".to_string(),
            },
        ));
    }

    // Валидация всех событий
    for (i, event) in events.iter().enumerate() {
        if let Err(err) = event.validate() {
            error!("Validation failed for event {}: {}", i, err);
            return Ok(json_response(
                StatusCode::BAD_REQUEST,
                &ErrorResponse {
                    error: format!("Validation failed: {}", err),
                },
            ));
        }
    }

    let mut errors = Vec::new();

    // Обработка каждого события
    for event in events.iter() {
        let queue_name = get_queue_name(event.is_system_event());

        match serde_json::to_vec(event) {
            Ok(event_json) => {
                if let Err(err) = state.rabbit_client.publish(queue_name, &event_json).await {
                    error!("Failed to publish event {}: {}", event.tx_id, err);
                    errors.push(EventError {
                        event: event.tx_id.clone(),
                        error: err.to_string(),
                    });
                }
            }
            Err(err) => {
                error!("Failed to serialize event {}: {}", event.tx_id, err);
                errors.push(EventError {
                    event: event.tx_id.clone(),
                    error: "Failed to serialize event".to_string(),
                });
            }
        }
    }

    if !errors.is_empty() {
        return Ok(json_response(
            StatusCode::BAD_REQUEST,
            &PartialSuccessResponse {
                status: "PARTIAL_SUCCESS".to_string(),
                processed: events.len() - errors.len(),
                errors,
            },
        ));
    }

    Ok(json_response(
        StatusCode::OK,
        &SuccessResponse {
            status: "SUCCESS".to_string(),
            processed: events.len(),
        },
    ))
}
