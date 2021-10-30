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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ContainerInjectorSpec defines the desired state of ContainerInjector
type ContainerInjectorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ContainerInjector. Edit containerinjector_types.go to remove/update
	Foo     string   `json:"foo,omitempty"`
	Image   string   `json:"image,omitempty"`
	Label   string   `json:"label,omitempty"`
	Command []string `json:"command,omitempty" protobuf:"bytes,3,rep,name=command"`
	Args    []string `json:"args,omitempty" protobuf:"bytes,4,rep,name=args"`
}

// ContainerInjectorStatus defines the observed state of ContainerInjector
type ContainerInjectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ContainerInjector is the Schema for the containerinjectors API
type ContainerInjector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerInjectorSpec   `json:"spec,omitempty"`
	Status ContainerInjectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContainerInjectorList contains a list of ContainerInjector
type ContainerInjectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerInjector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContainerInjector{}, &ContainerInjectorList{})
}
