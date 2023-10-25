/*
Copyright 2023 VMware Inc.

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
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/vmware-labs/reconciler-runtime/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MonoRepositorySpec defines the structure of the mono repository.
type MonoRepositorySpec struct {
	GitRepository v1beta2.GitRepositorySpec `json:"gitRepository"`
	Include       string                    `json:"include"`
}

// MonoRepositoryStatus defines the observed state of MonoRepository.
type MonoRepositoryStatus struct {
	apis.Status `json:",inline"`

	// URL is the dynamic fetch link for the latest Artifact.
	// It is provided on a "best effort" basis, and using the precise
	// GitRepositoryStatus.Artifact data is recommended.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the last successful GitRepository reconciliation.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	// ObservedInclude is the observed list of GitRepository resources used to
	// calculate the checksum for this artifact
	// +optional
	ObservedInclude string `json:"observedInclude,omitempty"`

	// ObservedFileList is the file list used to
	// calculate the checksum for this artifact
	// +optional
	ObservedFileList string `json:"observedFileList,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

// Artifact represents the output of a Source reconciliation.
type Artifact struct {
	// Path is the relative file path of the Artifact. It can be used to locate
	// the file in the root of the Artifact storage on the local file system of
	// the controller managing the Source.
	// +required
	Path string `json:"path"`

	// URL is the HTTP address of the Artifact as exposed by the controller
	// managing the Source. It can be used to retrieve the Artifact for
	// consumption, e.g. by another controller applying the Artifact contents.
	// +required
	URL string `json:"url"`

	// Revision is a human-readable identifier traceable in the origin source
	// system. It can be a Git commit SHA, Git tag, a Helm chart version, etc.
	// +optional
	Revision string `json:"revision"`

	// Checksum is the SHA256 checksum of the Artifact file.
	// Deprecated: use Artifact.Digest instead.
	// +optional
	Checksum string `json:"checksum,omitempty"`

	// Digest is the digest of the file in the form of '<algorithm>:<checksum>'.
	// +optional
	// +kubebuilder:validation:Pattern="^[a-z0-9]+(?:[.+_-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"
	Digest string `json:"digest,omitempty"`

	// LastUpdateTime is the timestamp corresponding to the last update of the
	// Artifact.
	// +required
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Size is the number of bytes in the file.
	// +optional
	Size *int64 `json:"size,omitempty"`

	// Metadata holds upstream information such as OCI annotations.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=monorepo
//+kubebuilder:printcolumn:name="Source Ref",type="string",JSONPath=`.spec.sourceRef.name`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
//+kubebuilder:printcolumn:name="Checksum",type="string",JSONPath=".status.artifact.checksum",description=""
//+kubebuilder:printcolumn:name="Last Update",type="date",JSONPath=".status.artifact.lastUpdateTime",description=""
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// MonoRepository is the Schema for the mono repository API.
type MonoRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonoRepositorySpec   `json:"spec,omitempty"`
	Status MonoRepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MonoRepositoryList contains a list of MonoRepository.
type MonoRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MonoRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MonoRepository{}, &MonoRepositoryList{})
}
