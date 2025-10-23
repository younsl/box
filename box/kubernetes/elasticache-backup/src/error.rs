use thiserror::Error;

#[derive(Error, Debug)]
pub enum BackupError {
    #[error("Snapshot not found: {0}")]
    NotFound(String),

    #[error("Snapshot creation failed: {0}")]
    SnapshotFailed(String),

    #[error("S3 export failed: {0}")]
    ExportFailed(String),

    #[error("Operation timed out: {0}")]
    Timeout(String),
}
