use std::env;

#[derive(Clone)]
pub struct Config {
    pub socket_path: String,
    pub rabbitmq_url: String,
}

impl Config {
    pub fn from_env() -> Self {
        let socket_path = env::var("SOCKET_PATH")
            .unwrap_or_else(|_| "/tmp/rust_actix/app.sock".to_string());
        
        let rabbitmq_url = env::var("DSN__RABBITMQ")
            .expect("DSN__RABBITMQ environment variable is required");

        Self {
            socket_path,
            rabbitmq_url,
        }
    }
}
