use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct ErrorData {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub code: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub message: Option<String>,
}

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct TrackData {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub priority: Option<i32>,
    pub is_system: bool,
}

#[derive(Debug, Deserialize, Serialize, Clone)]
pub struct StatusEvent {
    pub state: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<ErrorData>,
    #[serde(rename = "trackData")]
    #[serde(skip_serializing_if = "Option::is_none")]
    pub track_data: Option<TrackData>,
    #[serde(rename = "updatedAt")]
    pub updated_at: String,
    #[serde(rename = "txId")]
    pub tx_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub email: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub channel_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub channel: Option<String>,
}

impl StatusEvent {
    pub fn validate(&self) -> Result<(), String> {
        if self.state.is_empty() {
            return Err("field 'state' is required".to_string());
        }
        if self.updated_at.is_empty() {
            return Err("field 'updatedAt' is required".to_string());
        }
        if self.tx_id.is_empty() {
            return Err("field 'txId' is required".to_string());
        }
        Ok(())
    }

    pub fn is_system_event(&self) -> bool {
        self.track_data
            .as_ref()
            .map(|td| td.is_system)
            .unwrap_or(false)
    }
}
