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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!s
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BaselineSpec defines the desired state of Baseline
type BaselineSpec struct {
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=0
	// Cpu is the the number of cores
	Cpu *int32 `json:"cpu"`
	//+kubebuilder:validation:Optional
	// Memory is the amount of memory
	Memory string `json:"mem"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=0
	// Cpu is the the number of cores
	Io int32 `json:"io"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:validation:Minimum=0
	// Sock is the number of workers exercising socket I/O networking
	Sock int32 `json:"sock"`
	//+kubebuilder:validation:Optional
	// Custom is a custom string to pass to stress-ng
	Custom string `json:"custom"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="quay.io/jcastillolema/stressng:0.14.01"
	Image string `json:"image"`
	//+kubebuilder:validation:Optional
	HostNetwork bool `json:"hostNetwork"`
	//+kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector"`
	//+kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`
}

// BaselineStatus defines the observed state of Baseline
type BaselineStatus struct {
	Command string `json:"command"`
	Custom  string `json:"custom"`
}

//+kubebuilder:object:root=true

//+kubebuilder:printcolumn:name="Command",type=string,JSONPath=`.status.command`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
