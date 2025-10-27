use std::process::Command;

fn main() {
    // Get git commit hash
    // Try to get from git first, then fall back to GITHUB_SHA env var (for CI builds)
    let commit = std::env::var("GITHUB_SHA")
        .ok()
        .and_then(|sha| {
            if sha.len() >= 7 {
                Some(sha[0..7].to_string())
            } else {
                None
            }
        })
        .or_else(|| {
            Command::new("git")
                .args(["rev-parse", "--short", "HEAD"])
                .output()
                .ok()
                .and_then(|output| {
                    if output.status.success() {
                        String::from_utf8(output.stdout).ok()
                    } else {
                        None
                    }
                })
                .map(|s| s.trim().to_string())
        })
        .unwrap_or_else(|| "unknown".to_string());

    // Get build date
    let build_date = chrono::Utc::now().format("%Y-%m-%dT%H:%M:%SZ").to_string();

    // Set environment variables for compile-time embedding
    println!("cargo:rustc-env=GIT_COMMIT={}", commit);
    println!("cargo:rustc-env=BUILD_DATE={}", build_date);

    // Rerun if git HEAD changes (only if .git exists)
    if std::path::Path::new(".git/HEAD").exists() {
        println!("cargo:rerun-if-changed=.git/HEAD");
    }
}
