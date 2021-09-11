/*
Copyright 2021.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// QuarantineSpec defines the desired state of Quarantine
type QuarantineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Quarantine. Edit quarantine_types.go to remove/update
	Nodes     []Node     `json:"nodes,omitempty"`
	Debug     Debug      `json:"debug,omitempty"`
	Resources []Resource `json:"resources"`
}

type Node struct {
	Name    string `json:"name"`
	Isolate bool   `json:"isolate,omitempty"`
	Rescale bool   `json:"rescale,omitempty"`
}

type Resource struct {
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type Debug struct {
	Enabled   bool   `json:"enabled"`
	Image     string `json:"image,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// QuarantineStatus defines the observed state of Quarantine
type QuarantineStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Quarantine is the Schema for the quarantines API
type Quarantine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuarantineSpec   `json:"spec,omitempty"`
	Status QuarantineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuarantineList contains a list of Quarantine
type QuarantineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Quarantine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Quarantine{}, &QuarantineList{})
}
