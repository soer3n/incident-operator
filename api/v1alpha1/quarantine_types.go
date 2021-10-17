/*
Copyright 2021.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
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

// Node defines a configuration for node to isolate
type Node struct {
	Name      string     `json:"name"`
	Isolate   bool       `json:"isolate,omitempty"`
	Rescale   bool       `json:"rescale,omitempty"`
	Resources []Resource `json:"resources,omitempty"`
}

// Resource defines a workload to isolate on a node
type Resource struct {
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Keep      bool   `json:"keep,omitempty"`
}

// Debug defines a debug pod configuration
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
