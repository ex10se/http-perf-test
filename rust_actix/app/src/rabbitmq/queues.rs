pub const EXCHANGE_NAME: &str = "rust-actix";
pub const QUEUE_ACTIX: &str = "rust-actix";
pub const QUEUE_SYSTEM_ACTIX: &str = "system-rust-actix";

pub fn get_queue_name(is_system: bool) -> &'static str {
    if is_system {
        QUEUE_SYSTEM_ACTIX
    } else {
        QUEUE_ACTIX
    }
}
