# KServe

For KServe installation and usage, see the [GitHub Actions tests](.github/workflows/kserve_test.yaml) which demonstrate working configurations.

For complete documentation, visit the [official KServe website](https://kserve.github.io/website/).

## Integration with KubeFlow

When using KServe with path-based routing in a KubeFlow deployment, you may encounter VirtualService conflicts that result in 404 errors when accessing KServe InferenceServices.

**Common Issues:**
- KServe InferenceServices return 404 errors when accessed via their configured domain
- Conflicts between KubeFlow's wildcard VirtualServices and KServe's specific-host VirtualServices

**Solution:** See the [Istio troubleshooting guide](../../common/istio/README.md#virtualservice-conflicts-with-kserve-path-based-routing) for detailed resolution steps.

**Related Documentation:**
- [KServe Path-Based Routing Configuration](https://kserve.github.io/website/docs/admin-guide/configurations#path-template)
- [Upstream Istio Issue](https://github.com/istio/istio/issues/57404)