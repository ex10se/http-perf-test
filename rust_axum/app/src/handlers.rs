use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Json},
};
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

pub async fn status_handler(
    State(state): State<Arc<AppState>>,
    Json(events): Json<Vec<StatusEvent>>,
) -> impl IntoResponse {
    if events.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Request body must contain at least one event".to_string(),
            }),
        )
            .into_response();
    }

    // Валидация всех событий
    for (i, event) in events.iter().enumerate() {
        if let Err(err) = event.validate() {
            error!("Validation failed for event {}: {}", i, err);
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: format!("Validation failed: {}", err),
                }),
            )
                .into_response();
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
        return (
            StatusCode::BAD_REQUEST,
            Json(PartialSuccessResponse {
                status: "PARTIAL_SUCCESS".to_string(),
                processed: events.len() - errors.len(),
                errors,
            }),
        )
            .into_response();
    }

    (
        StatusCode::OK,
        Json(SuccessResponse {
            status: "SUCCESS".to_string(),
            processed: events.len(),
        }),
    )
        .into_response()
}
