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
	"github.com/vmware-labs/reconciler-runtime/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (in *SourceRef) Key(namespace string) client.ObjectKey {
	return client.ObjectKey{Name: in.Name, Namespace: namespace}
}

// SourceRef the GVK and name of the source object
type SourceRef struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	// FIXME namespace?
}

// FilterSpec defines the structure of the filter
type FilterSpec struct {
	SourceRef SourceRef `json:"sourceRef"`
}

// FilterStatus defines the observed state of Filter
type FilterStatus struct {
	apis.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Filter is the Schema for the filter API
type Filter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FilterSpec   `json:"spec,omitempty"`
	Status FilterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FilterList contains a list of Filter
type FilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Filter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Filter{}, &FilterList{})
}
