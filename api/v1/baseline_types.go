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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!s
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BaselineSpec defines the desired state of Baseline
type BaselineSpec struct {
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=1
	// Cpu is the the number of cores
	Cpu int32 `json:"cpu"`
	//+kubebuilder:validation:Optional
	// Memory is the ammount of memory
	Memory string `json:"memory"`
	//+kubebuilder:validation:Optional
	// Custom is a custom string to pass to stress-ng
	Custom string `json:"custom"`
}

// BaselineStatus defines the observed state of Baseline
type BaselineStatus struct {
	Command string `json:"command"`
}

//+kubebuilder:object:root=true

//+kubebuilder:subresource:status
// Baseline is the Schema for the baselines API
type Baseline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BaselineSpec   `json:"spec,omitempty"`
	Status BaselineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BaselineList contains a list of Baseline
type BaselineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Baseline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Baseline{}, &BaselineList{})
}
