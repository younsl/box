# EKS gp2 to gp3 Migration Guide

## Overview

This guide describes **three methods** for migrating EBS gp2 volumes to gp3 in Kubernetes environments. The first two methods enable in-place migration without requiring pod restart or PV recreation, allowing you to change volume types with zero downtime:

1. **Volume Attributes Class (VAC)** - Declarative approach (Recommended)
2. **PVC Annotation** - Imperative approach
3. **VolumeSnapshot** - Legacy approach (⚠️ requires pod restart)

## Volume Modification Process

**Recommended approach for production environments.** This section describes the modern, declarative method using VolumeAttributesClass (VAC) that enables zero-downtime, in-place volume migration with GitOps-friendly configuration management.

### Method 1: Volume Attributes Class (VAC) - Declarative Approach

This method performs in-place volume migration without requiring pod restart or PV recreation, enabling zero-downtime volume type changes. Best suited for GitOps workflows and managing multiple volumes with standardized profiles.

**Requirements for using VolumeAttributesClass:**
- [ ] Kubernetes 1.31 or later (VolumeAttributesClass promoted to Beta in upstream)
- [ ] EBS CSI driver v1.35.0+ (EKS managed add-on v1.35.0-eksbuild.2+)
- [ ] VolumeAttributesClass feature gate
  - EKS 1.31+: Automatically enabled by default
  - Upstream Kubernetes 1.31: Requires manual activation

![Kubernetes Architecture](./assets/gp3-migration-1.png)

#### What is VolumeAttributesClass?

[VolumeAttributesClass (VAC)](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/) defines a set of volume attributes (like type, IOPS, throughput) that can be applied to existing PersistentVolumes without recreating them. Think of it as a "profile" for volume modifications that enables zero-downtime in-place updates.

**Important**: The `parameters` field syntax varies by CSI driver. For `ebs.csi.aws.com`, use `type`, `iops`, and `throughput`. Other drivers may use different parameter names - always check your CSI driver's documentation.

#### Step 1: Create VolumeAttributesClass

Create [VolumeAttributesClass](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/) resource with desired gp3 configuration.

```bash
kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1beta1
kind: VolumeAttributesClass
metadata:
  name: gp3-migration
driverName: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
  tagSpecification_1: "performance=standard"
EOF
```

#### Step 2: Patch PVC

Patch existing PVC to reference the [VolumeAttributesClass](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/).

```bash
kubectl patch pvc <pvc-name> \
  --type merge \
  --patch '{"spec":{"volumeAttributesClassName":"gp3-migration"}}'
```

**Note**: `storageClassName` and `volumeAttributesClassName` can be different. The `storageClassName` (e.g., `gp2`) represents the original StorageClass used when the PV was **first created** and cannot be changed. The `volumeAttributesClassName` (e.g., `gp3-migration`) is used to **modify the existing PV's attributes** in-place. This is the expected pattern for volume migration scenarios.

#### References

- [AWS EBS CSI Driver - Modify Volume](https://github.com/kubernetes-sigs/aws-ebs-csi-driver/blob/master/docs/modify-volume.md)
- [AWS Blog - Modify Amazon EBS volumes on Kubernetes with Volume Attributes Classes](https://aws.amazon.com/ko/blogs/containers/modify-amazon-ebs-volumes-on-kubernetes-with-volume-attributes-classes/)

## Alternative Methods

**For quick migrations or older Kubernetes versions.** These methods provide alternatives when VolumeAttributesClass is not available or for simple one-off migrations. Method 2 supports in-place migration, while Method 3 uses snapshot-based restoration to create new volumes with pod restart.

<details>
<summary>Method 2: PVC Annotation - Quick Migration</summary>

This method also performs in-place volume migration without requiring pod restart or PV recreation.

Available for CSI driver **v1.19.0+**. This is the simplest approach for one-off migrations.

```bash
kubectl annotate pvc <pvc-name> ebs.csi.aws.com/volumeType="gp3"
```

</details>

<details>
<summary>Method 3: VolumeSnapshot - Legacy Approach</summary>

**⚠️ This method requires pod restart and creates new PV instead of modifying existing volumes in-place.**

While this outdated method works with any CSI driver version, it is no longer recommended because it requires pod restart and doesn't support in-place migration. **Use VAC or PVC Annotation instead for zero-downtime migrations.**

**Recommended approach**: Use **[VAC](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/)** for declarative infrastructure management, **PVC Annotation** for quick migrations, or **VolumeSnapshot** only when you need backup guarantees during migration.

</details>

## Migration Method Comparison

| Aspect | [Volume Attributes Class](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/) | PVC Annotation Method | VolumeSnapshot Method |
|--------|------------------------|----------------------|----------------------|
| **Complexity** | Moderate (3-4 steps) | Simple (1 command) | Complex (5+ steps) |
| **Downtime** | None | None | Requires pod restart |
| **Existing Volumes** | Migrates in-place | Migrates in-place | Creates new volumes |
| **K8s Version** | 1.31+ | Any | Any |
| **CSI Driver Required** | v1.35.0+ | v1.19.0+ | Any version |
| **Risk Level** | Low | Low | Medium |
| **Management Style** | Declarative (GitOps) | Imperative (ad-hoc) | Manual |
| **Reusability** | High (profiles) | Low | Low |
| **Best For** | Infrastructure as Code | Quick migrations | Backup during migration |
| **Approach** | Create VAC → Patch PVC | `kubectl annotate pvc` | Create snapshot → Restore |
| **Guide** | [AWS Blog](https://aws.amazon.com/ko/blogs/containers/modify-amazon-ebs-volumes-on-kubernetes-with-volume-attributes-classes/) | [AWS re:Post](https://repost.aws/knowledge-center/eks-migrate-ebs-volume-g3) | [AWS Blog](https://aws.amazon.com/ko/blogs/containers/migrating-amazon-eks-clusters-from-gp2-to-gp3-ebs-volumes/) |

## Additional Resources

- [Kubernetes Storage Volume Attributes Classes](https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/)
- [Migrating Amazon EKS clusters from gp2 to gp3 EBS volumes](https://aws.amazon.com/ko/blogs/containers/migrating-amazon-eks-clusters-from-gp2-to-gp3-ebs-volumes/)
