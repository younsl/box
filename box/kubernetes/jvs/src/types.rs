use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Serialize, Deserialize)]
pub struct PodList {
    pub items: Vec<Pod>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Pod {
    pub metadata: PodMetadata,
    #[serde(default, rename = "ownerReferences")]
    pub owner_references: Vec<OwnerReference>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PodMetadata {
    pub name: String,
    pub namespace: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct OwnerReference {
    pub kind: String,
}

impl Pod {
    pub fn is_daemonset(&self) -> bool {
        self.owner_references
            .iter()
            .any(|owner| owner.kind == "DaemonSet")
    }
}

#[derive(Debug)]
pub struct NamespaceStats {
    pub total_pods: usize,
    pub jdk_pods: usize,
}

#[derive(Debug)]
pub struct ScanResult {
    pub total_pods: usize,
    pub jdk_pods: usize,
    pub pod_versions: HashMap<String, HashMap<String, String>>, // namespace -> pod -> version
    pub namespace_stats: HashMap<String, NamespaceStats>, // namespace -> stats
}

impl ScanResult {
    pub fn new() -> Self {
        Self {
            total_pods: 0,
            jdk_pods: 0,
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
