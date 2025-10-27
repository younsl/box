use anyhow::Result;
use tracing::{error, info};

use ec2_statuscheck_rebooter::config::Config;
use ec2_statuscheck_rebooter::health::HealthServer;
use ec2_statuscheck_rebooter::logging;
use ec2_statuscheck_rebooter::rebooter::StatusCheckRebooter;

const HEALTH_CHECK_PORT: u16 = 8080;

#[tokio::main]
async fn main() -> Result<()> {
    let config = Config::from_args();
    logging::init(&config.log_format, &config.log_level);

    info!(
        version = env!("CARGO_PKG_VERSION"),
        commit = env!("GIT_COMMIT"),
        build_date = env!("BUILD_DATE"),
        "EC2 Status Check Rebooter starting"
    );

    // Start health check server
    let start_time = std::time::Instant::now();
    let health_server = HealthServer::new();
    let health_server_clone = health_server.clone();

    let (tx, rx) = tokio::sync::oneshot::channel();
    tokio::spawn(async move {
        match health_server_clone.serve(HEALTH_CHECK_PORT, tx).await {
            Ok(_) => {}
            Err(e) => {
                error!(
                    error = %e,
                    "Health check server failed"
                );
            }
        }
    });

    // Wait for health server to be ready
    rx.await.ok();
    let startup_time_ms = start_time.elapsed().as_millis();
    info!(
        startup_time_ms = startup_time_ms,
        "Health check server initialization complete"
    );

    let mut rebooter = match StatusCheckRebooter::new(config, health_server.clone()).await {
        Ok(r) => r,
        Err(e) => {
            error!(
                error = %e,
                "Failed to initialize rebooter"
            );
            std::process::exit(1);
        }
    };

    // Handle graceful shutdown
    tokio::select! {
        result = rebooter.run() => {
            if let Err(e) = result {
                error!(
                    error = %e,
                    "Rebooter encountered a fatal error"
                );
                std::process::exit(1);
            }
        }
        _ = tokio::signal::ctrl_c() => {
            info!("Received SIGINT/SIGTERM, initiating graceful shutdown");
            info!("Shutdown complete");
        }
    }

    Ok(())
}
