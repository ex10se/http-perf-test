pub const EXCHANGE_NAME: &str = "rust-axum";
pub const QUEUE_AXUM: &str = "rust-axum";
pub const QUEUE_SYSTEM_AXUM: &str = "system-rust-axum";

pub fn get_queue_name(is_system: bool) -> &'static str {
    if is_system {
        QUEUE_SYSTEM_AXUM
    } else {
        QUEUE_AXUM
    }
}
