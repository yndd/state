/*
Copyright 2022 NDD.

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
	"reflect"

	nddv1 "github.com/yndd/ndd-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A StateSpec defines the desired state of a State.
type StateSpec struct {
	nddv1.ResourceSpec `json:",inline"`
	//+kubebuilder:pruning:PreserveUnknownFields
	//+kubebuilder:validation:Required
	Properties runtime.RawExtension `json:"properties,omitempty"`
}

// A StateStatus represents the observed state of a State.
type StateStatus struct {
	nddv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// State is the Schema for the State API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TARGET",type="string",JSONPath=".status.conditions[?(@.kind=='TargetFound')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.conditions[?(@.kind=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNC",type="string",JSONPath=".status.conditions[?(@.kind=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories={ndd,nddp}
type State struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StateSpec   `json:"spec,omitempty"`
	Status StateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StateList contains a list of State
type StateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []State `json:"items"`
}

func init() {
	SchemeBuilder.Register(&State{}, &StateList{})
}

// State type metadata.
var (
	StateKind         = reflect.TypeOf(State{}).Name()
	StateGroupKind        = schema.GroupKind{Group: Group, Kind: StateKind}.String()
	StateKindAPIVersion   = StateKind + "." + GroupVersion.String()
	StateGroupVersionKind = GroupVersion.WithKind(StateKind)
)
