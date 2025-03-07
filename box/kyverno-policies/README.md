# kyverno-policies

## Summary

Collection of Kyverno policies for Kubernetes cluster security and governance. These policies are tested and verified in production environments.

## Best Practices for Enterprise-grade Kubernetes

For enterprise-grade Kubernetes clusters using [Kyverno](https://github.com/kyverno/kyverno) as a policy engine, it is **highly recommended** to manage all policy custom resources using the official [kyverno-policies](https://github.com/kyverno/kyverno/tree/main/charts/kyverno-policies) helm chart provided by Kyverno. This approach ensures consistent deployment, versioning, and maintenance of security policies across your infrastructure.

Managing all resources through helm charts rather than directly applying YAML manifests with `kubectl` is a fundamental best practice for Kubernetes resource management. This practice provides significant advantages in versioning, rollback capabilities, templating, and maintaining configuration consistency across multiple clusters.

## References

For various sample policies, visit the [Kyverno Official Policies][kyverno-policies] page.

[kyverno-policies]: https://kyverno.io/policies