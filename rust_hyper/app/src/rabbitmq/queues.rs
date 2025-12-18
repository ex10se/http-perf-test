pub const EXCHANGE_NAME: &str = "rust-hyper";
pub const QUEUE_HYPER: &str = "rust-hyper";
pub const QUEUE_SYSTEM_HYPER: &str = "system-rust-hyper";

pub fn get_queue_name(is_system: bool) -> &'static str {
    if is_system {
        QUEUE_SYSTEM_HYPER
    } else {
        QUEUE_HYPER
    }
}
