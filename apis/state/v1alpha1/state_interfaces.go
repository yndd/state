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
	nddov1 "github.com/yndd/nddo-runtime/apis/common/v1"
)

/*
var _ IFStateList = &StateList{}

// +k8s:deepcopy-gen=false
type IFStateList interface {
	client.ObjectList

	GetDevices() []IFState
}

func (x *StateList) GetDevices() []IFState {
	xs := make([]IFState, len(x.Items))
	for i, r := range x.Items {
		r := r // Pin range variable so we can take its address.
		xs[i] = &r
	}
	return xs
}

var _ IFState = &State{}

// +k8s:deepcopy-gen=false
type IFState interface {
	resource.Object
	resource.Conditioned

	GetDeploymentPolicy() nddv1.DeploymentPolicy
	SetDeploymentPolicy(b nddv1.DeploymentPolicy)
	GetDeletionPolicy() nddv1.DeletionPolicy
	SetDeletionPolicy(r nddv1.DeletionPolicy)
	GetHierPaths() map[string][]string
	SetHierPaths(n map[string][]string)
	GetTargetReference() *nddv1.Reference
	SetTargetReference(r *nddv1.Reference)
	GetRootPaths() []string
	SetRootPaths(n []string)

	GetCondition(ct nddv1.ConditionKind) nddv1.Condition
	SetConditions(c ...nddv1.Condition)
	// getters based on labels
	GetOwner() string
	//GetDeploymentPolicy() string
	GetDeviceName() string
	GetEndpointGroup() string
	GetOrganization() string
	GetDeployment() string
	GetAvailabilityZone() string
	// Spec
	GetSpec() *StateSpec
}
*/

func (x *State) GetOwner() string {
	if s, ok := x.GetLabels()[nddov1.LabelNddaOwner]; !ok {
		return ""
	} else {
		return s
	}
}

func (x *State) GetDeviceName() string {
	if s, ok := x.GetLabels()[nddov1.LabelNddaDevice]; !ok {
		return ""
	} else {
		return s
	}
}

func (x *State) GetEndpointGroup() string {
	if s, ok := x.GetLabels()[nddov1.LabelNddaEndpointGroup]; !ok {
		return ""
	} else {
		return s
	}
}

func (x *State) GetOrganization() string {
	if s, ok := x.GetLabels()[nddov1.LabelNddaOrganization]; !ok {
		return ""
	} else {
		return s
	}
}

func (x *State) GetDeployment() string {
	if s, ok := x.GetLabels()[nddov1.LabelNddaDeployment]; !ok {
		return ""
	} else {
		return s
	}
}

func (x *State) GetAvailabilityZone() string {
	if s, ok := x.GetLabels()[nddov1.LabelNddaAvailabilityZone]; !ok {
		return ""
	} else {
		return s
	}
}

func (x *State) GetSpec() *StateSpec {
	return &x.Spec
}
