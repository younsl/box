package scanner

import "time"

// PodList represents the Kubernetes API response for listing pods
type PodList struct {
	Items []Pod `json:"items"`
}

// Pod represents a Kubernetes pod
type Pod struct {
	Metadata        PodMetadata      `json:"metadata"`
	OwnerReferences []OwnerReference `json:"ownerReferences,omitempty"`
}

// PodMetadata contains pod metadata
type PodMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// OwnerReference contains owner reference information
type OwnerReference struct {
	Kind string `json:"kind"`
}

// ScanResult contains the results of a pod scan
type ScanResult struct {
	TotalPods   int
	JDKPods     int
	ElapsedTime time.Duration
	PodVersions map[string]map[string]string // namespace -> pod -> version
}

// PodVersion represents a single pod's Java version information
type PodVersion struct {
	Namespace string
	PodName   string
	Version   string
}
