use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Represents a semantic version (x.y.z) with optional build number
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Version {
    pub major: u32,
    pub minor: u32,
    pub patch: u32,
    pub build: u32, // For Java 8 versions like 1.8.0_292
}

impl Version {
    /// Parse version string like "11.0.16", "1.8.0_292", "17", "20.3.0"
    pub fn parse(version_str: &str) -> Option<Self> {
        if version_str == "Unknown" {
            return None;
        }

        // Split by underscore to handle Java 8 versions like "1.8.0_292"
        let parts: Vec<&str> = version_str.split('_').collect();
        let version_part = parts[0];
        let build_part = parts.get(1);

        // Parse build number if exists (e.g., "292" from "1.8.0_292")
        let build = build_part
            .and_then(|b| b.parse::<u32>().ok())
            .unwrap_or(0);

        // Parse the main version part (e.g., "1.8.0" from "1.8.0_292")
        let version_components: Vec<&str> = version_part.split('.').collect();

        match version_components.len() {
            1 => {
                // "17" -> 17.0.0.0
                let major = version_components[0].parse().ok()?;
                Some(Version {
                    major,
                    minor: 0,
                    patch: 0,
                    build,
                })
            }
            2 => {
                // "20.3" -> 20.3.0.0
                let major = version_components[0].parse().ok()?;
                let minor = version_components[1].parse().ok()?;
                Some(Version {
                    major,
                    minor,
                    patch: 0,
                    build,
                })
            }
            3 => {
                // "11.0.16" or "1.8.0" -> 11.0.16.0 or 1.8.0.0
                let major = version_components[0].parse().ok()?;
                let minor = version_components[1].parse().ok()?;
                let patch = version_components[2].parse().ok()?;
                Some(Version {
                    major,
                    minor,
                    patch,
                    build,
                })
            }
            _ => None,
        }
    }

    /// Compare if this version is less than another version
    pub fn is_less_than(&self, other: &Version) -> bool {
        if self.major != other.major {
            return self.major < other.major;
        }
        if self.minor != other.minor {
            return self.minor < other.minor;
        }
        if self.patch != other.patch {
            return self.patch < other.patch;
        }
        self.build < other.build
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PodList {
    pub items: Vec<Pod>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Pod {
    pub metadata: PodMetadata,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PodMetadata {
    pub name: String,
    pub namespace: String,
    #[serde(default, rename = "ownerReferences")]
    pub owner_references: Vec<OwnerReference>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct OwnerReference {
    pub kind: String,
}

impl Pod {
    pub fn is_daemonset(&self) -> bool {
        self.metadata.owner_references
            .iter()
            .any(|owner| owner.kind == "DaemonSet")
    }
}

#[derive(Debug)]
pub struct NamespaceStats {
    pub total_pods: usize,
    pub jdk_pods: usize,
    pub node_pods: usize,
}

#[derive(Debug)]
pub struct PodVersion {
    pub java: String,
    pub node: String,
}

impl PodVersion {
    pub fn has_java(&self) -> bool {
        self.java != "Unknown"
    }

    pub fn has_node(&self) -> bool {
        self.node != "Unknown"
    }

    /// Check if Java version is below the minimum threshold
    pub fn java_below_min(&self, min_version: &Version) -> bool {
        if let Some(current_version) = Version::parse(&self.java) {
            current_version.is_less_than(min_version)
        } else {
            false
        }
    }

    /// Check if Node version is below the minimum threshold
    pub fn node_below_min(&self, min_version: &Version) -> bool {
        if let Some(current_version) = Version::parse(&self.node) {
            current_version.is_less_than(min_version)
        } else {
            false
        }
    }
}

#[derive(Debug)]
pub struct ScanResult {
    pub total_pods: usize,
    pub jdk_pods: usize,
    pub node_pods: usize,
    pub pod_versions: HashMap<String, HashMap<String, PodVersion>>, // namespace -> pod -> versions
    pub namespace_stats: HashMap<String, NamespaceStats>, // namespace -> stats
}

impl ScanResult {
    pub fn new() -> Self {
        Self {
            total_pods: 0,
            jdk_pods: 0,
            node_pods: 0,
            pod_versions: HashMap::new(),
            namespace_stats: HashMap::new(),
        }
    }
}

impl Default for ScanResult {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version_parse() {
        // Test single digit version
        let v = Version::parse("17").unwrap();
        assert_eq!(v, Version { major: 17, minor: 0, patch: 0, build: 0 });

        // Test x.y version
        let v = Version::parse("20.3").unwrap();
        assert_eq!(v, Version { major: 20, minor: 3, patch: 0, build: 0 });

        // Test x.y.z version
        let v = Version::parse("11.0.16").unwrap();
        assert_eq!(v, Version { major: 11, minor: 0, patch: 16, build: 0 });

        // Test Java 8 versions with underscore and build number
        let v = Version::parse("1.8.0_292").unwrap();
        assert_eq!(v, Version { major: 1, minor: 8, patch: 0, build: 292 });

        let v = Version::parse("1.8.0_232").unwrap();
        assert_eq!(v, Version { major: 1, minor: 8, patch: 0, build: 232 });

        let v = Version::parse("1.8.0_342").unwrap();
        assert_eq!(v, Version { major: 1, minor: 8, patch: 0, build: 342 });

        // Test Unknown
        assert!(Version::parse("Unknown").is_none());
    }

    #[test]
    fn test_version_comparison() {
        let v15 = Version::parse("15").unwrap();
        let v11 = Version::parse("11").unwrap();
        let v17 = Version::parse("17.0.2").unwrap();
        let v1_8 = Version::parse("1.8.0_292").unwrap();

        // Test less than
        assert!(v11.is_less_than(&v15));
        assert!(v1_8.is_less_than(&v11));
        assert!(!v17.is_less_than(&v15));
        assert!(!v15.is_less_than(&v15));

        // Test Java 8 build number comparison
        let v1_8_232 = Version::parse("1.8.0_232").unwrap();
        let v1_8_292 = Version::parse("1.8.0_292").unwrap();
        let v1_8_342 = Version::parse("1.8.0_342").unwrap();

        // Same major.minor.patch, different build numbers
        assert!(v1_8_232.is_less_than(&v1_8_292));
        assert!(v1_8_292.is_less_than(&v1_8_342));
        assert!(!v1_8_342.is_less_than(&v1_8_232));

        // Test Node versions
        let v20_3_0 = Version::parse("20.3.0").unwrap();
        let v18_17_0 = Version::parse("18.17.0").unwrap();
        let v20_5_1 = Version::parse("20.5.1").unwrap();

        assert!(v18_17_0.is_less_than(&v20_3_0));
        assert!(!v20_5_1.is_less_than(&v20_3_0));
        assert!(v20_3_0.is_less_than(&v20_5_1));
    }

    #[test]
    fn test_pod_version_filtering() {
        let pod_version = PodVersion {
            java: "11.0.16".to_string(),
            node: "18.17.0".to_string(),
        };

        let java_min = Version::parse("15").unwrap();
        let node_min = Version::parse("20.3.0").unwrap();

        // Java 11 is below 15
        assert!(pod_version.java_below_min(&java_min));

        // Node 18.17.0 is below 20.3.0
        assert!(pod_version.node_below_min(&node_min));

        // Test higher versions
        let pod_version2 = PodVersion {
            java: "17.0.2".to_string(),
            node: "20.5.1".to_string(),
        };

        // Java 17 is NOT below 15
        assert!(!pod_version2.java_below_min(&java_min));

        // Node 20.5.1 is NOT below 20.3.0
        assert!(!pod_version2.node_below_min(&node_min));
    }
}
