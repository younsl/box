# gp3-migration

## Summary

All gp2 type EBS volumes located in the specified AWS Region are converted to gp3.

## Precautions

- Each EBS volume can only be modified once **every 6 hours**.
- See [AWS EBS volume modification documentation](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-modify-volume.html#elastic-volumes-considerations) for detailed requirements and limitations.

## Kubernetes Volume Impact

This script migrates **all gp2 volumes** in the specified region, including volumes used by Kubernetes PersistentVolumes (PVs).

- **Online migration**: EBS volume type changes can be performed without detaching the volume (no downtime)
- **Kubernetes compatibility**: StorageClass name (e.g., `gp2`) is just a label and doesn't need to match the actual EBS volume type
- **Existing PVs unaffected**: Already provisioned PVs remain bound; volume type change is transparent to Kubernetes
- **Performance impact**: Temporary I/O performance degradation may occur during migration
- **Database workloads**: For production databases, run during low-traffic periods
- **Recommended approach**: Migrate in phases (dev → staging → production stateless → production stateful)

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
