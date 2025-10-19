use std::process::Command;

fn main() {
    // Get git commit hash
    let output = Command::new("git")
        .args(["rev-parse", "--short", "HEAD"])
        .output();

    let git_hash = match output {
        Ok(output) if output.status.success() => {
            String::from_utf8(output.stdout).unwrap_or_else(|_| "unknown".to_string())
        }
        _ => "unknown".to_string(),
    };

    // Remove trailing newline
    let git_hash = git_hash.trim();

    // Set environment variable for compile time
    println!("cargo:rustc-env=GIT_HASH={}", git_hash);
}
