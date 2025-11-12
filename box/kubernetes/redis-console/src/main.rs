mod config;
mod redis_client;
mod tty;

use anyhow::{Context, Result};
use clap::Parser;
use colored::Colorize;
use config::Config;
use redis_client::RedisClient;
use rustyline::DefaultEditor;
use rustyline::error::ReadlineError;
use std::path::PathBuf;
use tabled::{Table, Tabled};
use tokio::time::{Duration, timeout};
use tty::tty_writeln;

#[derive(Parser, Debug)]
#[command(
    name = "redis-console",
    version,
    about = "Interactive CLI for managing multiple Redis clusters"
)]
struct Args {
    /// Configuration file path
    #[arg(short, long)]
    config: Option<PathBuf>,
}

#[derive(Tabled)]
struct ClusterRow {
    #[tabled(rename = "ID")]
    id: usize,
    #[tabled(rename = "Alias")]
    alias: String,
    #[tabled(rename = "Host")]
    host: String,
    #[tabled(rename = "Port")]
    port: u16,
    #[tabled(rename = "Engine")]
    engine: String,
    #[tabled(rename = "Version")]
    version: String,
    #[tabled(rename = "Mode")]
    mode: String,
    #[tabled(rename = "TLS")]
    tls: String,
    #[tabled(rename = "Status")]
    status: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Args::parse();

    // Load configuration
    let config = if let Some(config_path) = args.config {
        Config::load_from_file(&config_path)
            .with_context(|| format!("Failed to load config from {:?}", config_path))?
    } else {
        Config::load().context("Failed to load configuration")?
    };

    let _ = tty_writeln(&format!(
        "{}",
        r#"
    ____           ___          ______                       __
   / __ \___  ____/ (_)____    / ____/___  ____  _________  / /__
  / /_/ / _ \/ __  / / ___/   / /   / __ \/ __ \/ ___/ __ \/ / _ \
 / _, _/  __/ /_/ / (__  )   / /___/ /_/ / / / (__  ) /_/ / /  __/
/_/ |_|\___/\__,_/_/____/    \____/\____/_/ /_/____/\____/_/\___/
    "#
        .bright_cyan()
        .bold()
    ));
    let _ = tty_writeln(&format!(
        "{}",
        format!(
            "Interactive CLI for managing multiple Redis clusters - v{}",
            env!("CARGO_PKG_VERSION")
        )
        .bright_cyan()
    ));
    let _ = tty_writeln("");

    // Run REPL
    run_repl(config).await
}

async fn run_repl(config: Config) -> Result<()> {
    let mut rl = DefaultEditor::new()?;
    let mut current_cluster: Option<(usize, RedisClient)> = None;

    // Check if TTY is available
    let tty_status = if std::fs::OpenOptions::new()
        .write(true)
        .open("/dev/tty")
        .is_ok()
    {
        "tty"
    } else {
        "no-tty"
    };

    print_help();
    let _ = tty_writeln("");

    loop {
        let prompt = if let Some((idx, _)) = &current_cluster {
            format!(
                "{}.redis-console [{}]> ",
                tty_status, config.clusters[*idx].alias
            )
        } else {
            format!("{}.redis-console> ", tty_status)
        };

        let readline = rl.readline(&prompt);

        match readline {
            Ok(line) => {
                let line = line.trim();
                if line.is_empty() {
                    continue;
                }

                rl.add_history_entry(line)?;

                let parts: Vec<&str> = line.split_whitespace().collect();
                if parts.is_empty() {
                    continue;
                }

                let command = parts[0].to_lowercase();

                match command.as_str() {
                    "help" | "h" => print_help(),
                    "list" | "ls" | "l" => check_health(&config).await,
                    "connect" | "c" => {
                        if parts.len() < 2 {
                            let _ = tty_writeln(&format!(
                                "{}",
                                "Usage: connect <cluster-id or name>".red()
                            ));
                        } else {
                            current_cluster = connect_cluster(&config, parts[1]).await;
                        }
                    }
                    "info" => {
                        if let Some((_, ref mut client)) = current_cluster {
                            execute_info(client).await;
                        } else {
                            let _ = tty_writeln(&format!(
                                "{}",
                                "Not connected. Use 'connect <id>' first".red()
                            ));
                        }
                    }
                    "quit" | "exit" | "q" => {
                        if current_cluster.is_some() {
                            // Disconnect if connected
                            let (idx, _) = current_cluster.take().unwrap();
                            let _ = tty_writeln(&format!(
                                "{}",
                                format!("Disconnected from {}", config.clusters[idx].alias)
                                    .yellow()
                            ));
                        } else {
                            // Quit if not connected
                            let _ = tty_writeln(&format!("{}", "Goodbye!".bright_cyan()));
                            break;
                        }
                    }
                    _ => {
                        // Execute as Redis command
                        if let Some((_, ref mut client)) = current_cluster {
                            execute_redis_command(client, line).await;
                        } else {
                            let _ = tty_writeln(&format!(
                                "{}",
                                "Unknown command. Type 'help' for available commands".red()
                            ));
                        }
                    }
                }
            }
            Err(ReadlineError::Interrupted) => {
                let _ = tty_writeln(&format!("{}", "CTRL-C".yellow()));
                continue;
            }
            Err(ReadlineError::Eof) => {
                let _ = tty_writeln(&format!("{}", "CTRL-D".yellow()));
                break;
            }
            Err(err) => {
                // Critical error - keep in stdout for kubectl logs
                println!("{}", format!("Error: {:?}", err).red());
                break;
            }
        }

        let _ = tty_writeln("");
    }

    Ok(())
}

fn print_help() {
    let _ = tty_writeln(&format!("{}", "Available Commands:".bright_yellow().bold()));
    let _ = tty_writeln(&format!(
        "  {}  - Show this help message",
        "help, h".green()
    ));
    let _ = tty_writeln(&format!(
        "  {}  - List all clusters with health status",
        "list, ls, l".green()
    ));
    let _ = tty_writeln(&format!(
        "  {}  - Connect to a cluster by ID or name",
        "connect <id|name>, c".green()
    ));
    let _ = tty_writeln(&format!(
        "  {}  - Show Redis server info (when connected)",
        "info".green()
    ));
    let _ = tty_writeln(&format!(
        "  {}  - Disconnect (if connected) or exit",
        "quit, exit, q".green()
    ));
    let _ = tty_writeln("");
    let _ = tty_writeln(&format!(
        "{}",
        "When connected, you can execute any Redis command directly:".bright_yellow()
    ));
    let _ = tty_writeln("  Example: GET mykey");
    let _ = tty_writeln("  Example: SET mykey value");
    let _ = tty_writeln("  Example: KEYS *");
}

async fn check_health(config: &Config) {
    let _ = tty_writeln(&format!(
        "{}",
        format!(
            "Checking {} clusters from {}",
            config.clusters.len(),
            config.source_description()
        )
        .bright_cyan()
    ));

    let mut rows = Vec::new();

    for (i, cluster) in config.clusters.iter().enumerate() {
        let (status, engine, version, mode) = match timeout(
            Duration::from_secs(2),
            RedisClient::connect(cluster.clone()),
        )
        .await
        {
            Ok(Ok(mut client)) => {
                // Get engine, version and mode info
                match client.get_server_info().await {
                    Ok((e, v, m)) => ("Healthy".to_string(), e, v, m),
                    Err(_) => (
                        "Healthy".to_string(),
                        "unknown".to_string(),
                        "unknown".to_string(),
                        "unknown".to_string(),
                    ),
                }
            }
            Ok(Err(e)) => (
                format!("Unhealthy: {}", e),
                "-".to_string(),
                "-".to_string(),
                "-".to_string(),
            ),
            Err(_) => (
                "Timeout".to_string(),
                "-".to_string(),
                "-".to_string(),
                "-".to_string(),
            ),
        };

        rows.push(ClusterRow {
            id: i,
            alias: cluster.alias.clone(),
            host: cluster.host.clone(),
            port: cluster.port,
            engine,
            version,
            mode,
            tls: if cluster.tls { "Yes" } else { "No" }.to_string(),
            status,
        });
    }

    let table = Table::new(rows).to_string();
    let _ = tty_writeln(&table);
}

async fn connect_cluster(config: &Config, identifier: &str) -> Option<(usize, RedisClient)> {
    // Try to parse as ID
    let cluster_idx = if let Ok(id) = identifier.parse::<usize>() {
        if id < config.clusters.len() {
            Some(id)
        } else {
            None
        }
    } else {
        // Try to find by alias
        config
            .clusters
            .iter()
            .position(|c| c.alias.eq_ignore_ascii_case(identifier))
    };

    match cluster_idx {
        Some(idx) => {
            let cluster = &config.clusters[idx];
            let _ = tty_writeln(&format!(
                "{}",
                format!("Connecting to {}...", cluster.alias).bright_cyan()
            ));

            match RedisClient::connect(cluster.clone()).await {
                Ok(client) => {
                    let _ = tty_writeln(&format!(
                        "{}",
                        format!("Connected to {}", cluster.alias).green()
                    ));
                    Some((idx, client))
                }
                Err(e) => {
                    let _ = tty_writeln(&format!("{}", format!("Failed to connect: {}", e).red()));
                    None
                }
            }
        }
        None => {
            let _ = tty_writeln(&format!(
                "{}",
                format!("Cluster '{}' not found", identifier).red()
            ));
            None
        }
    }
}

async fn execute_info(client: &mut RedisClient) {
    match client.info().await {
        Ok(info) => {
            let _ = tty_writeln(&format!("{}", "Redis Server Info:".bright_yellow().bold()));
            let _ = tty_writeln(&info);
        }
        Err(e) => {
            let _ = tty_writeln(&format!("{}", format!("Error getting info: {}", e).red()));
        }
    }
}

async fn execute_redis_command(client: &mut RedisClient, command: &str) {
    match client.execute_command(command).await {
        Ok(output) => {
            if output.is_empty() {
                let _ = tty_writeln(&format!("{}", "(nil)".bright_black()));
            } else {
                let _ = tty_writeln(&output);
            }
        }
        Err(e) => {
            let _ = tty_writeln(&format!("{}", format!("Error: {}", e).red()));
        }
    }
}
