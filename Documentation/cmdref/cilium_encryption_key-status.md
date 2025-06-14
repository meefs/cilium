<!-- This file was autogenerated via cilium cmdref, do not edit manually-->

## cilium encryption key-status

Display IPsec key

### Synopsis

This command displays IPsec encryption key

```
cilium encryption key-status [flags]
```

### Options

```
  -h, --help                     help for key-status
  -o, --output string            Output format. One of: json, summary (default "summary")
      --wait-duration duration   Maximum time to wait for result, default 1 minute (default 1m0s)
```

### Options inherited from parent commands

```
      --as string                  Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray       Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --context string             Kubernetes configuration context
      --helm-release-name string   Helm release name (default "cilium")
      --kubeconfig string          Path to the kubeconfig file
  -n, --namespace string           Namespace Cilium is running in (default "kube-system")
```

### SEE ALSO

* [cilium encryption](cilium_encryption.md)	 - Cilium encryption

