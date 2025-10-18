use anyhow::{Context, Result};
use serde::Deserialize;
use std::fs;
use std::path::Path;

#[derive(Debug, Deserialize)]
pub struct Config {
    pub domains: Vec<String>,
}

impl Config {
    /// Load configuration from a YAML file
    pub fn load<P: AsRef<Path>>(path: P) -> Result<Self> {
        let path = path.as_ref();

        // Read file content
        let content = fs::read_to_string(path)
            .with_context(|| format!("Failed to read config file: {}", path.display()))?;

        // Parse YAML
        let config: Config = serde_yaml::from_str(&content)
            .with_context(|| format!("Failed to parse YAML from {}", path.display()))?;

        // Validate
        if config.domains.is_empty() {
            anyhow::bail!("No domains found in config file");
        }

        Ok(config)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;
    use tempfile::NamedTempFile;

    #[test]
    fn test_load_valid_config() {
        let mut temp_file = NamedTempFile::new().unwrap();
        writeln!(temp_file, "domains:").unwrap();
        writeln!(temp_file, "  - example.com").unwrap();
        writeln!(temp_file, "  - https://test.com").unwrap();

        let config = Config::load(temp_file.path()).unwrap();
        assert_eq!(config.domains.len(), 2);
        assert_eq!(config.domains[0], "example.com");
    }

    #[test]
    fn test_load_empty_domains() {
        let mut temp_file = NamedTempFile::new().unwrap();
        writeln!(temp_file, "domains: []").unwrap();

        let result = Config::load(temp_file.path());
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("No domains"));
    }
}
