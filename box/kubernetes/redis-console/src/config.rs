use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// Redis cluster configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClusterConfig {
    pub alias: String,
    pub host: String,
    #[serde(default = "default_port")]
    pub port: u16,
    #[serde(default)]
    pub password: Option<String>,
    #[serde(default)]
    pub tls: bool,
    #[serde(default)]
    pub cluster_mode: bool,
    #[serde(default)]
    pub description: Option<String>,
}

fn default_port() -> u16 {
    6379
}

impl ClusterConfig {
    /// Get connection URL for redis client
    pub fn connection_url(&self) -> String {
        let scheme = if self.tls { "rediss" } else { "redis" };
        let auth = self
            .password
            .as_ref()
            .map(|p| format!(":{p}@"))
            .unwrap_or_default();

        format!("{scheme}://{auth}{}:{}", self.host, self.port)
    }
}

/// Application configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    #[serde(default)]
    pub clusters: Vec<ClusterConfig>,
    #[serde(default)]
    pub aws_region: Option<String>,
    #[serde(skip)]
    pub source: ConfigSource,
}

/// Configuration source
#[derive(Debug, Clone, Default)]
pub enum ConfigSource {
    File(String),
    #[default]
    Empty,
}

impl Config {
    /// Load configuration from file
    pub fn load_from_file(path: &PathBuf) -> Result<Self> {
        let contents = std::fs::read_to_string(path)
            .with_context(|| format!("Failed to read config file: {}", path.display()))?;

        let mut config: Self = serde_yaml::from_str(&contents)
            .with_context(|| format!("Failed to parse config file: {}", path.display()))?;

        // Convert to absolute path for clarity
        let absolute_path = std::fs::canonicalize(path)
            .unwrap_or_else(|_| path.clone())
            .display()
            .to_string();

        config.source = ConfigSource::File(absolute_path);
        Ok(config)
    }

    /// Load configuration from default locations
    pub fn load() -> Result<Self> {
        // Try /etc/redis/clusters/config.yaml first (container path)
        let container_path = PathBuf::from("/etc/redis/clusters/config.yaml");
        if container_path.exists() {
            return Self::load_from_file(&container_path);
        }

        // Try ~/.config/redis-console/config.yaml
        if let Some(config_dir) = dirs::config_dir() {
            let user_path = config_dir.join("redis-console").join("config.yaml");
            if user_path.exists() {
                return Self::load_from_file(&user_path);
            }
        }

        // Return empty config if no file found
        Ok(Self {
            clusters: Vec::new(),
            aws_region: None,
            source: ConfigSource::Empty,
        })
    }

    /// Get a human-readable source description
    pub fn source_description(&self) -> String {
        match &self.source {
            ConfigSource::File(path) => format!("file: {}", path),
            ConfigSource::Empty => "empty (no config file found)".to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_connection_url() {
        let config = ClusterConfig {
            alias: "test".to_string(),
            host: "localhost".to_string(),
            port: 6379,
            password: None,
            tls: false,
            cluster_mode: false,
            description: None,
        };

        assert_eq!(config.connection_url(), "redis://localhost:6379");
    }

    #[test]
    fn test_connection_url_with_password() {
        let config = ClusterConfig {
            alias: "test".to_string(),
            host: "localhost".to_string(),
            port: 6379,
            password: Some("secret".to_string()),
            tls: false,
            cluster_mode: false,
            description: None,
        };

        assert_eq!(config.connection_url(), "redis://:secret@localhost:6379");
    }

    #[test]
    fn test_connection_url_with_tls() {
        let config = ClusterConfig {
            alias: "test".to_string(),
            host: "localhost".to_string(),
            port: 6380,
            password: None,
            tls: true,
            cluster_mode: false,
            description: None,
        };

        assert_eq!(config.connection_url(), "rediss://localhost:6380");
    }
}
