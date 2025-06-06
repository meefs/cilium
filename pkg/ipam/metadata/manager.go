// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package metadata

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/cilium/statedb"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/cilium/cilium/daemon/k8s"
	"github.com/cilium/cilium/pkg/annotation"
	"github.com/cilium/cilium/pkg/ipam"
	"github.com/cilium/cilium/pkg/logging/logfields"
)

type ManagerStoppedError struct{}

func (m *ManagerStoppedError) Error() string {
	return "ipam-metadata-manager has been stopped"
}

type ResourceNotFound struct {
	Resource  string
	Name      string
	Namespace string
}

func (r *ResourceNotFound) Error() string {
	name := r.Name
	if r.Namespace != "" {
		name = r.Namespace + "/" + r.Name
	}
	return fmt.Sprintf("resource %s %q not found", r.Resource, name)
}

func (r *ResourceNotFound) Is(target error) bool {
	targetErr, ok := target.(*ResourceNotFound)
	if !ok {
		return false
	}
	if r != nil && targetErr.Resource != "" {
		return r.Resource == targetErr.Resource
	}
	return true
}

type Manager interface {
	GetIPPoolForPod(owner string, family ipam.Family) (pool string, err error)
}

type manager struct {
	logger     *slog.Logger
	db         *statedb.DB
	pods       statedb.Table[k8s.LocalPod]
	namespaces statedb.Table[k8s.Namespace]
}

func splitK8sPodName(owner string) (namespace, name string, ok bool) {
	// Require namespace/name format
	namespace, name, ok = strings.Cut(owner, "/")
	if !ok {
		return "", "", false
	}
	// Check if components are a valid namespace name and pod name
	if validation.IsDNS1123Subdomain(namespace) != nil ||
		validation.IsDNS1123Subdomain(name) != nil {
		return "", "", false
	}
	return namespace, name, true
}

func determinePoolByAnnotations(annotations map[string]string, family ipam.Family) (pool string, ok bool) {
	switch family {
	case ipam.IPv4:
		if annotations[annotation.IPAMIPv4PoolKey] != "" {
			return annotations[annotation.IPAMIPv4PoolKey], true
		} else if annotations[annotation.IPAMPoolKey] != "" {
			return annotations[annotation.IPAMPoolKey], true
		}
	case ipam.IPv6:
		if annotations[annotation.IPAMIPv6PoolKey] != "" {
			return annotations[annotation.IPAMIPv6PoolKey], true
		} else if annotations[annotation.IPAMPoolKey] != "" {
			return annotations[annotation.IPAMPoolKey], true
		}
	}

	return "", false
}

func (m *manager) GetIPPoolForPod(owner string, family ipam.Family) (pool string, err error) {
	if family != ipam.IPv6 && family != ipam.IPv4 {
		return "", fmt.Errorf("invalid IP family: %s", family)
	}

	namespace, name, ok := splitK8sPodName(owner)
	if !ok {
		m.logger.Debug(
			"IPAM metadata request for invalid pod name, falling back to default pool",
			logfields.Owner, owner,
		)
		return ipam.PoolDefault().String(), nil
	}

	txn := m.db.ReadTxn()

	// Check annotation on pod
	pod, _, found := m.pods.Get(txn, k8s.PodByName(namespace, name))
	if !found {
		return "", &ResourceNotFound{Resource: "Pod", Namespace: namespace, Name: name}
	} else if ippool, ok := determinePoolByAnnotations(pod.Annotations, family); ok {
		return ippool, nil
	}

	// Check annotation on namespace
	podNamespace, _, found := m.namespaces.Get(txn, k8s.NamespaceIndex.Query(namespace))
	if !found {
		return "", &ResourceNotFound{Resource: "Namespace", Name: namespace}
	} else if ippool, ok := determinePoolByAnnotations(podNamespace.Annotations, family); ok {
		return ippool, nil
	}

	// Fallback to default pool
	return ipam.PoolDefault().String(), nil
}
