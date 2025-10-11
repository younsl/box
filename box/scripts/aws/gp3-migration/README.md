# gp3-migration

## Summary

All gp2 type EBS volumes located in the specified AWS Region are converted to gp3.

## EKS gp2 to gp3 Migration

For migrating EBS volumes in Kubernetes clusters, see the dedicated guide: **[EKS gp2 to gp3 Migration Guide](/box/til/eks-gp2-to-gp3-migration.md)**

The guide covers three migration methods. The first two methods support in-place migration without requiring pod restart or PV recreation, enabling zero-downtime volume type changes:
- **Volume Attributes Class (VAC)** - Declarative approach for K8s 1.31+ (Recommended)
- **PVC Annotation** - Quick imperative approach for CSI driver v1.19.0+
- **VolumeSnapshot** - Legacy approach (⚠️ requires pod restart, not recommended)

**This script is still useful for**:
- Migrating standalone EBS volumes not managed by Kubernetes
- Bulk migration of existing gp2 volumes across multiple AWS accounts/regions
- Environments using older versions of aws-ebs-csi-driver

## Precautions

- Each EBS volume can only be modified once **every 6 hours**.
- **Online migration**: EBS volume type changes can be performed without detaching the volume (no downtime)
- **Performance impact**: Temporary I/O performance degradation may occur during migration
- **Database workloads**: For production databases, run during low-traffic periods
- See [AWS EBS volume modification documentation](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-modify-volume.html#elastic-volumes-considerations) for detailed requirements and limitations.

## Example

```bash
export AWS_PROFILE=dev
sh gp3_migration.sh
```

```bash
[i] Start finding all gp2 volumes in ap-northeast-2
[i] List up all gp2 volumes in ap-northeast-2
=========================================
vol-1234567890abcdef0
vol-0987654321abcdef0
vol-abcdefgh123456780
vol-ijklmnop123456780
vol-12345678abcdefgh0
vol-098765abcdef12340
vol-abcdef12345678900
=========================================
Do you want to proceed with the migration? (y/n): y
[i] Starting volume migration...
[i] Migrating all gp2 volumes to gp3
[i] Volume vol-1234567890abcdef0 changed to state 'modifying' successfully.
[i] Volume vol-0987654321abcdef0 changed to state 'modifying' successfully.
[i] Volume vol-abcdefgh123456780 changed to state 'modifying' successfully.
[i] Volume vol-ijklmnop123456780 changed to state 'modifying' successfully.
[i] Volume vol-12345678abcdefgh0 changed to state 'modifying' successfully.
[i] Volume vol-098765abcdef12340 changed to state 'modifying' successfully.
[i] Volume vol-abcdef12345678900 changed to state 'modifying' successfully.
[i] All gp2 volumes have been migrated to gp3 successfully!
```

## References

- [Blog post](https://younsl.github.io/blog/script-gp2-volumes-to-gp3-migration/)
- [AWS EBS volume modification documentation](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-modify-volume.html#elastic-volumes-considerations)
- [Migrating Amazon EKS clusters from gp2 to gp3 EBS volumes](https://aws.amazon.com/ko/blogs/containers/migrating-amazon-eks-clusters-from-gp2-to-gp3-ebs-volumes/)
