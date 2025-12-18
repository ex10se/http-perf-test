use actix_web::{web, HttpResponse, Responder};
use log::error;
use serde::{Deserialize, Serialize};
use std::sync::Arc;

use crate::models::StatusEvent;
use crate::rabbitmq::{get_queue_name, RabbitClient};

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

pub struct AppState {
    pub rabbit_client: RabbitClient,
}

pub async fn status_handler(
    events: web::Json<Vec<StatusEvent>>,
    data: web::Data<Arc<AppState>>,
) -> impl Responder {
    if events.is_empty() {
        return HttpResponse::BadRequest().json(ErrorResponse {
            error: "Request body must contain at least one event".to_string(),
        });
    }

    // Валидация всех событий
    for (i, event) in events.iter().enumerate() {
        if let Err(err) = event.validate() {
            error!("Validation failed for event {}: {}", i, err);
            return HttpResponse::BadRequest().json(ErrorResponse {
                error: format!("Validation failed: {}", err),
            });
        }
    }

    let mut errors = Vec::new();

    // Обработка каждого события
    for event in events.iter() {
        let queue_name = get_queue_name(event.is_system_event());

        match serde_json::to_vec(event) {
            Ok(event_json) => {
                if let Err(err) = data
                    .rabbit_client
                    .publish(queue_name, &event_json)
                    .await
                {
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
        return HttpResponse::BadRequest().json(PartialSuccessResponse {
            status: "PARTIAL_SUCCESS".to_string(),
            processed: events.len() - errors.len(),
            errors,
        });
    }

    HttpResponse::Ok().json(SuccessResponse {
        status: "SUCCESS".to_string(),
        processed: events.len(),
    })
}
