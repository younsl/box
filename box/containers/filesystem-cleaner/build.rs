use std::env;
use std::process::Command;

fn main() {
    // Get git commit hash - prefer env var, fallback to git command
    let git_hash = env::var("VERGEN_GIT_SHA").unwrap_or_else(|_| {
        Command::new("git")
            .args(["rev-parse", "--short", "HEAD"])
            .output()
            .ok()
            .and_then(|output| String::from_utf8(output.stdout).ok())
            .unwrap_or_else(|| "unknown".to_string())
            .trim()
            .to_string()
    });

    // Get build timestamp - prefer env var, fallback to current time
    let timestamp = env::var("VERGEN_BUILD_TIMESTAMP").unwrap_or_else(|_| {
        chrono::Utc::now()
            .format("%Y-%m-%d %H:%M:%S UTC")
            .to_string()
    });

    println!("cargo:rustc-env=VERGEN_GIT_SHA={}", git_hash);
    println!("cargo:rustc-env=VERGEN_BUILD_TIMESTAMP={}", timestamp);
    println!("cargo:rerun-if-changed=.git/HEAD");
    println!("cargo:rerun-if-env-changed=VERGEN_GIT_SHA");
    println!("cargo:rerun-if-env-changed=VERGEN_BUILD_TIMESTAMP");
}
