/*
Copyright 2022.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NginxOpSpec defines the desired state of NginxOp
type NginxOpSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Replicas int32  `json:"replicas,omitempty"`
	Host     string `json:"host,omitempty"`
	Image    string `json:"image,omitempty"`
}

// NginxOpStatus defines the observed state of NginxOp
type NginxOpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	URL string `json:"url"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NginxOp is the Schema for the nginxops API
type NginxOp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxOpSpec   `json:"spec,omitempty"`
	Status NginxOpStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NginxOpList contains a list of NginxOp
type NginxOpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxOp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NginxOp{}, &NginxOpList{})
}
