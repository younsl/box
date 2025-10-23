use anyhow::{Context, Result};
use clap::Parser;
use image::Luma;
use qrcode::QrCode;
use std::path::PathBuf;

const VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Parser)]
#[command(
    name = "qg",
    version = VERSION,
    about = "QR code generator",
    long_about = "A simple QR code generator that creates a QR code from a given URL."
)]
struct Cli {
    /// URL to encode in the QR code (must start with http:// or https://)
    #[arg(value_name = "URL")]
    url: String,

    /// Output filename for the QR code
    #[arg(short, long, default_value = "qrcode.png")]
    filename: PathBuf,

    /// Width of the QR code in pixels
    #[arg(long, default_value = "100")]
    width: u32,

    /// Height of the QR code in pixels
    #[arg(long, default_value = "100")]
    height: u32,

    /// Suppress output messages
    #[arg(short, long)]
    quiet: bool,
}

fn validate_url(url: &str) -> Result<()> {
    if !url.starts_with("http://") && !url.starts_with("https://") {
        anyhow::bail!(
            "Invalid URL. The URL must start with http:// or https://. Please check the URL and try again."
        );
    }
    Ok(())
}

fn generate_qr_code(url: &str, filename: &PathBuf, width: u32, height: u32) -> Result<()> {
    // Generate QR code
    let code = QrCode::new(url.as_bytes())
        .context("Failed to generate QR code")?;

    // Render to image
    let image = code.render::<Luma<u8>>()
        .min_dimensions(width, height)
        .build();

    // Save as PNG
    image.save(filename)
        .with_context(|| format!("Failed to save QR code to {}", filename.display()))?;

    Ok(())
}

fn main() -> Result<()> {
    let cli = Cli::parse();

    // Validate URL
    validate_url(&cli.url)?;

    // Generate QR code
    generate_qr_code(&cli.url, &cli.filename, cli.width, cli.height)?;

    // Print success message
    if !cli.quiet {
        println!("QR code saved as {}.", cli.filename.display());
        println!("Address: {}. Size: {}x{}", cli.url, cli.width, cli.height);
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_url_valid_http() {
        assert!(validate_url("http://example.com").is_ok());
    }

    #[test]
    fn test_validate_url_valid_https() {
        assert!(validate_url("https://example.com").is_ok());
    }

    #[test]
    fn test_validate_url_invalid() {
        assert!(validate_url("example.com").is_err());
        assert!(validate_url("ftp://example.com").is_err());
    }
}
