use anyhow::Result;
use bytesize::ByteSize;
use std::fs;
use std::path::{Path, PathBuf};
use tracing::{info, warn};

use crate::matcher::PatternMatcher;

/// Information about a file to be cleaned
#[derive(Debug)]
pub struct FileInfo {
    pub path: PathBuf,
    pub size: u64,
}

/// File system scanner for collecting files based on patterns
///
/// Responsible for traversing directories and collecting files
/// that match the configured patterns.
pub struct FileScanner<'a> {
    matcher: &'a PatternMatcher,
}

impl<'a> FileScanner<'a> {
    /// Create a new file scanner with a pattern matcher
    pub fn new(matcher: &'a PatternMatcher) -> Self {
        Self { matcher }
    }

    /// Scan a directory and collect all matching files
    pub fn scan(&self, base_path: &Path) -> Vec<FileInfo> {
        let mut files = Vec::new();

        match self.walk_directory(base_path, base_path, &mut files) {
            Ok(_) => {}
            Err(e) => {
                warn!(
                    path = %base_path.display(),
                    error = %e,
                    "Error walking directory"
                );
            }
        }

        files
    }

    /// Recursively walk a directory tree and collect matching files
    fn walk_directory(
        &self,
        base_path: &Path,
        current_dir: &Path,
        files: &mut Vec<FileInfo>,
    ) -> Result<()> {
        if !current_dir.exists() {
            return Ok(());
        }

        let entries = match fs::read_dir(current_dir) {
            Ok(entries) => entries,
            Err(e) => {
                warn!(path = %current_dir.display(), error = %e, "Error reading directory");
                return Ok(());
            }
        };

        for entry in entries {
            let entry = match entry {
                Ok(e) => e,
                Err(e) => {
                    warn!(error = %e, "Error reading directory entry");
                    continue;
                }
            };

            let path = entry.path();

            // Get relative path from base_path for glob matching
            let relative_path = match path.strip_prefix(base_path) {
                Ok(rel) => rel.to_string_lossy().to_string(),
                Err(_) => {
                    // Fallback to file name if strip_prefix fails
                    match path.file_name() {
                        Some(name) => name.to_string_lossy().to_string(),
                        None => continue,
                    }
                }
            };

            // Use symlink_metadata instead of metadata to detect symlinks
            // metadata() follows symlinks, symlink_metadata() does not
            let metadata = match std::fs::symlink_metadata(&path) {
                Ok(m) => m,
                Err(e) => {
                    warn!(path = %path.display(), error = %e, "Error reading metadata");
                    continue;
                }
            };

            // Skip symbolic links to prevent infinite loops and unintended deletions outside target-paths
            if metadata.is_symlink() {
                // Try to read the symlink target for logging
                let target = match std::fs::read_link(&path) {
                    Ok(t) => format!("{}", t.display()),
                    Err(_) => String::from("(target unreadable)"),
                };

                info!(
                    symlink = %path.display(),
                    relative_path = %relative_path,
                    target = %target,
                    file_type = "symlink",
                    "Skipping symbolic link to prevent infinite loops and unintended deletions outside target-paths"
                );
                continue;
            }

            if metadata.is_dir() {
                // Check if directory should be excluded using relative path
                if self.matcher.should_exclude(&relative_path) {
                    info!(
                        dir = %path.display(),
                        relative_path = %relative_path,
                        file_type = "directory",
                        "Skipping excluded directory"
                    );
                    continue;
                }
                // Recursively walk subdirectory
                let _ = self.walk_directory(base_path, &path, files);
            } else {
                // Process file using relative path for pattern matching
                if self.matcher.should_exclude(&relative_path) {
                    info!(
                        file = %path.display(),
                        relative_path = %relative_path,
                        file_type = "file",
                        size = %ByteSize::b(metadata.len()),
                        "Skipping excluded file"
                    );
                    continue;
                }

                if !self.matcher.should_include(&relative_path) {
                    continue;
                }

                files.push(FileInfo {
                    path,
                    size: metadata.len(),
                });
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::matcher::PatternMatcher;
    use std::fs::File;
    use std::io::Write;
    use tempfile::TempDir;

    fn create_test_file(base: &Path, path: &str, content: &[u8]) {
        let full_path = base.join(path);
        if let Some(parent) = full_path.parent() {
            fs::create_dir_all(parent).unwrap();
        }
        File::create(full_path).unwrap().write_all(content).unwrap();
    }

    #[test]
    fn test_scan_with_exclude() {
        let temp_dir = TempDir::new().unwrap();
        let temp_path = temp_dir.path();

        create_test_file(temp_path, "test.txt", b"test");
        create_test_file(temp_path, ".git/config", b"config");
        create_test_file(temp_path, "node_modules/lib.js", b"js");

        let matcher = PatternMatcher::new(
            &["*".to_string()],
            &["**/.git/**".to_string(), "**/node_modules/**".to_string()],
        )
        .unwrap();

        let scanner = FileScanner::new(&matcher);
        let files = scanner.scan(temp_path);

        assert_eq!(files.len(), 1);
        assert!(files.iter().any(|f| f.path.ends_with("test.txt")));
    }

    #[test]
    fn test_scan_nested_directories() {
        let temp_dir = TempDir::new().unwrap();
        let temp_path = temp_dir.path();

        create_test_file(temp_path, "build/groovy-dsl/cache.jar", b"jar");
        create_test_file(temp_path, "build/other/file.txt", b"txt");

        let matcher =
            PatternMatcher::new(&["*".to_string()], &["**/groovy-dsl/**".to_string()]).unwrap();

        let scanner = FileScanner::new(&matcher);
        let files = scanner.scan(temp_path);

        assert_eq!(files.len(), 1);
        assert!(files.iter().any(|f| f.path.ends_with("file.txt")));
    }

    #[test]
    fn test_skip_symbolic_links() {
        let temp_dir = TempDir::new().unwrap();
        let temp_path = temp_dir.path();

        // Create a real file
        create_test_file(temp_path, "real_file.txt", b"content");

        // Create a target directory for symlink
        create_test_file(temp_path, "target/important.dat", b"important");

        // Create symbolic link
        #[cfg(unix)]
        {
            use std::os::unix::fs::symlink;
            let link_path = temp_path.join("link_to_target");
            let target_path = temp_path.join("target");
            let _ = symlink(&target_path, &link_path);
        }

        let matcher = PatternMatcher::new(&["*".to_string()], &[]).unwrap();
        let scanner = FileScanner::new(&matcher);
        let files = scanner.scan(temp_path);

        // Should find real_file.txt and target/important.dat
        // Should NOT traverse through link_to_target
        assert!(files.iter().any(|f| f.path.ends_with("real_file.txt")));
        assert!(files.iter().any(|f| f.path.ends_with("important.dat")));

        // Count should be 2 (real_file.txt, target/important.dat)
        // NOT 3 (which would include target accessed via symlink)
        assert_eq!(files.len(), 2);
    }

    #[test]
    #[cfg(unix)]
    fn test_skip_circular_symlinks() {
        let temp_dir = TempDir::new().unwrap();
        let temp_path = temp_dir.path();

        use std::os::unix::fs::symlink;

        // Create a directory
        fs::create_dir(temp_path.join("dir")).unwrap();
        create_test_file(temp_path, "dir/file.txt", b"test");

        // Create circular symlink: dir/link_to_parent -> ..
        let link_path = temp_path.join("dir/link_to_parent");
        let _ = symlink("..", &link_path);

        let matcher = PatternMatcher::new(&["*".to_string()], &[]).unwrap();
        let scanner = FileScanner::new(&matcher);
        let files = scanner.scan(temp_path);

        // Should complete without infinite loop
        // Should find dir/file.txt only
        assert_eq!(files.len(), 1);
        assert!(files.iter().any(|f| f.path.ends_with("file.txt")));
    }
}
