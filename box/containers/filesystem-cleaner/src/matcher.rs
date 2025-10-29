use anyhow::Result;
use globset::{Glob, GlobSet, GlobSetBuilder};

/// Pattern matcher for include/exclude file patterns
///
/// Responsible for determining whether a file path matches
/// the configured glob patterns.
pub struct PatternMatcher {
    include_matcher: GlobSet,
    exclude_matcher: GlobSet,
}

impl PatternMatcher {
    /// Create a new pattern matcher from include and exclude patterns
    pub fn new(include_patterns: &[String], exclude_patterns: &[String]) -> Result<Self> {
        let include_matcher = Self::build_matcher(include_patterns)?;
        let exclude_matcher = Self::build_matcher(exclude_patterns)?;

        Ok(Self {
            include_matcher,
            exclude_matcher,
        })
    }

    /// Build a GlobSet from a list of patterns
    fn build_matcher(patterns: &[String]) -> Result<GlobSet> {
        let mut builder = GlobSetBuilder::new();
        for pattern in patterns {
            builder.add(Glob::new(pattern)?);
        }
        Ok(builder.build()?)
    }

    /// Check if a path should be excluded based on exclude patterns
    pub fn should_exclude(&self, path: &str) -> bool {
        self.exclude_matcher.is_match(path)
    }

    /// Check if a path matches include patterns
    pub fn should_include(&self, path: &str) -> bool {
        self.include_matcher.is_match(path)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_exclude_patterns() {
        let matcher = PatternMatcher::new(
            &["*".to_string()],
            &["**/.git/**".to_string(), "**/node_modules/**".to_string()],
        )
        .unwrap();

        assert!(matcher.should_exclude("project1/.git/config"));
        assert!(matcher.should_exclude("src/node_modules/lib.js"));
        assert!(!matcher.should_exclude("src/main.rs"));
    }

    #[test]
    fn test_include_patterns() {
        let matcher = PatternMatcher::new(&["*.txt".to_string()], &[]).unwrap();

        assert!(matcher.should_include("file.txt"));
        assert!(matcher.should_include("readme.txt"));
        assert!(!matcher.should_include("file.rs"));
    }

    #[test]
    fn test_nested_glob_patterns() {
        let matcher =
            PatternMatcher::new(&["*".to_string()], &["**/groovy-dsl/**".to_string()]).unwrap();

        assert!(matcher.should_exclude("build/groovy-dsl/cache.jar"));
        assert!(matcher.should_exclude("a/b/c/groovy-dsl/file.txt"));
        assert!(!matcher.should_exclude("build/other/file.jar"));
    }

    #[test]
    fn test_simple_filename_pattern() {
        let matcher = PatternMatcher::new(&["*".to_string()], &["app.log".to_string()]).unwrap();

        assert!(matcher.should_exclude("app.log"));
        assert!(!matcher.should_exclude("project1/app.log"));
    }

    #[test]
    fn test_file_extension_pattern() {
        let matcher = PatternMatcher::new(&["*".to_string()], &["*.log".to_string()]).unwrap();

        assert!(matcher.should_exclude("app.log"));
        assert!(matcher.should_exclude("project1/debug.log"));
        assert!(!matcher.should_exclude("app.txt"));
    }
}
