/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RBACSpec defines the RBAC configuration for the namespace
type RBACSpec struct {
	// admins is a list of users with admin access to the namespace
	// +kubebuilder:validation:MinItems=1
	Admins []string `json:"admins"`
	// viewers is a list of users with read-only access to the namespace
	// +optional
	Viewers []string `json:"viewers,omitempty"`
}

// ResourceQuotaSpec defines the resource limits for the namespace
type ResourceQuotaSpec struct {
	// cpu is the maximum CPU allocation for the namespace
	// +kubebuilder:validation:Required
	CPU string `json:"cpu"`
	// memory is the maximum memory allocation for the namespace
	// +kubebuilder:validation:Required
	Memory string `json:"memory"`
}

// ManagedNamespaceSpec defines the desired state of ManagedNamespace
type ManagedNamespaceSpec struct {
	// team is the name of the team that owns this namespace
	// +kubebuilder:validation:Required
	Team string `json:"team"`

	// environment is the deployment environment (dev, staging, prod)
	// +kubebuilder:validation:Enum=dev;staging;prod
	Environment string `json:"environment"`

	// resourceQuota defines CPU and memory limits for the namespace
	// +kubebuilder:validation:Required
	ResourceQuota ResourceQuotaSpec `json:"resourceQuota"`

	// rbac defines admin and viewer access for the namespace
	// +kubebuilder:validation:Required
	RBAC RBACSpec `json:"rbac"`
}

// ManagedNamespaceStatus defines the observed state of ManagedNamespace
type ManagedNamespaceStatus struct {
	// conditions represent the current state of the ManagedNamespace resource
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// namespaceName is the name of the provisioned namespace
	// +optional
	NamespaceName string `json:"namespaceName,omitempty"`

	// phase is the current phase of the namespace provisioning
	// +optional
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ManagedNamespace is the Schema for the managednamespaces API
type ManagedNamespace struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of ManagedNamespace
	// +required
	Spec ManagedNamespaceSpec `json:"spec"`

	// status defines the observed state of ManagedNamespace
	// +optional
	Status ManagedNamespaceStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ManagedNamespaceList contains a list of ManagedNamespace
type ManagedNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []ManagedNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedNamespace{}, &ManagedNamespaceList{})
}
