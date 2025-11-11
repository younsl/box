use anyhow::{Context, Result};
use redis::aio::ConnectionManager;
use std::time::Duration;
use tokio::time::timeout;

use crate::config::ClusterConfig;

/// Redis client wrapper
pub struct RedisClient {
    manager: ConnectionManager,
}

impl RedisClient {
    /// Create a new Redis client connection
    pub async fn connect(config: ClusterConfig) -> Result<Self> {
        let client = redis::Client::open(config.connection_url())
            .context("Failed to create Redis client")?;

        let manager = timeout(Duration::from_secs(5), ConnectionManager::new(client))
            .await
            .context("Connection timeout")??;

        Ok(Self { manager })
    }

    /// Execute INFO command
    pub async fn info(&mut self) -> Result<String> {
        let info: String = redis::cmd("INFO")
            .query_async(&mut self.manager)
            .await
            .context("Failed to execute INFO command")?;

        Ok(info)
    }

    /// Get Redis server version and mode from INFO
    pub async fn get_server_info(&mut self) -> Result<(String, String)> {
        let info: String = redis::cmd("INFO")
            .arg("server")
            .query_async(&mut self.manager)
            .await
            .context("Failed to execute INFO server command")?;

        let mut version = "unknown".to_string();
        let mut mode = "standalone".to_string();

        for line in info.lines() {
            if line.starts_with("redis_version:") {
                version = line
                    .split(':')
                    .nth(1)
                    .unwrap_or("unknown")
                    .trim()
                    .to_string();
            } else if line.starts_with("redis_mode:") {
                mode = line
                    .split(':')
                    .nth(1)
                    .unwrap_or("standalone")
                    .trim()
                    .to_string();
            }
        }

        Ok((version, mode))
    }

    /// Execute custom command
    pub async fn execute_command(&mut self, cmd: &str) -> Result<String> {
        let parts: Vec<&str> = cmd.split_whitespace().collect();
        if parts.is_empty() {
            return Ok(String::new());
        }

        let mut redis_cmd = redis::cmd(parts[0]);
        for arg in &parts[1..] {
            redis_cmd.arg(*arg);
        }

        let result: redis::Value = redis_cmd
            .query_async(&mut self.manager)
            .await
            .context("Failed to execute command")?;

        Ok(format_redis_value(&result))
    }
}

/// Format Redis value for display
fn format_redis_value(value: &redis::Value) -> String {
    match value {
        redis::Value::Nil => "(nil)".to_string(),
        redis::Value::Int(i) => format!("(integer) {}", i),
        redis::Value::BulkString(bytes) => String::from_utf8_lossy(bytes).to_string(),
        redis::Value::Array(arr) => {
            let items: Vec<String> = arr.iter().map(format_redis_value).collect();
            items.join("\n")
        }
        redis::Value::SimpleString(s) => s.clone(),
        redis::Value::Okay => "OK".to_string(),
        redis::Value::Double(d) => format!("(double) {}", d),
        // Catch-all for other variants
        _ => format!("{:?}", value),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_format_redis_value() {
        assert_eq!(format_redis_value(&redis::Value::Nil), "(nil)");
        assert_eq!(format_redis_value(&redis::Value::Int(42)), "(integer) 42");
        assert_eq!(
            format_redis_value(&redis::Value::SimpleString("OK".to_string())),
            "OK"
        );
    }
}
